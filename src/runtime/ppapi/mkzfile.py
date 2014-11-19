#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""This is a code generator.  It reads configuration information that
describes types, functions, and enums; and it uses the info to expand
code templates.  The config contains lines of the following form:

Enumeration: <type> is the Go type name, and <pattern> is a regular expression
for finding constants in the PPAPI header files.

  //enum <type> <pattern>

For example, the following defines a type Error containing all constants
matching the pattern PP_ERROR_.

  //enum Error PP_ERROR_

Types: define a type <type> with a machine representation <mtype>.  The machine
types includes the usual integer types int32, int64; floats float32 and float64;
pointers *; and structs with size in bytes struct[16].

  //type <type> <mtype>

  //type pp_Bool int32
  //type pp_Var  struct[16]

Callbacks: specify a callback with the given parameters.

  //callback f(a1 t1, ..., aN, tN) returnType

  //callback ppp_did_create(instance pp_Instance, argc uint32, argn **byte, argv **byte) pp_Bool

Functions: specify a function call with the given parameters.  This is a call
into PPAPI.

  //func f(a1 t1, ..., aN, tN) returnType = interface[index]

  //func ppb_var_from_utf8(data *byte, len uint32) pp_Var = PPB_VAR[2]

To run this, you will need a copy of the NaCl SDK, downloadable from
https://developer.chrome.com/native-client/sdk/download.

Let $NACL_SDK be the location of the NaCl SDK, and $PEPPER be the version of
pepper (e.g. pepper_34).  This script scans the header files in that package and
performs a template expansion.  The following command can be used to extract the
PPAPI constants.

./mkzfile.py --include=$NACL_SDK/$PEPPER/include/ppapi/c consts.got > zconsts.go
"""

__author__ = 'jyh@google.com (Jason Hickey)'

import argparse
import fileinput
import glob
import re
import string
import sys
from jinja2 import Environment, Template

parser = argparse.ArgumentParser(description='Process some integers.')
parser.add_argument('--include', nargs=1, help='PPAPI include directory from the NaCl SDK')
parser.add_argument('template', nargs='+', help='template files')

# Sizes for machine kinds.
sizes = {
    'void': (0, 0),
    'int32': (4, 4),
    'int64': (8, 8),
    'float32': (4, 4),
    'float64': (8, 8),
    '*': (4, 4),
    }

# align aligns an address to the next larger boundary.
def align(addr, n):
  return (addr + n - 1) & ~(n - 1)

# sizeof returns a triple of the kind name, the number of bytes needed to
# represent a value of that kind, and the alignment.
def sizeof(kind):
  if kind.startswith('*'):
    return ('*', 4, 4)
  elif kind.startswith('struct'):
    m = re.search(r'struct\[(\w+)(,(\w+))?\]', kind)
    if m == None:
      raise Exception("Malformed struct type: " + kind)
    a = 4
    if m.group(3) != None:
      a = int(m.group(3))
    return ('struct', int(m.group(1)), a)
  else:
    n, a = sizes[kind]
    return (kind, n, a)

# Type represents a type, where <name> is the name of the type, <kind> is the
# machine kind, and <size> is the sizeof that type.
class Type:
  """Type represents base type"""
  def __init__(self, name, kind, builtin=False):
    self.name = name
    self.kind, self.size, self.align = sizeof(kind)
    self.builtin = builtin

# Arg represents a function parameter, including the name and the type.
class Arg:
  """Argument"""
  def __init__(self, n, ty):
    self.name = n
    self.type = ty

# Func represents a function declaration, declared in the input file.
class Func:
  """Func represents a function description"""
  def __init__(self, name, args, result, interface, index):
    self.name = name
    self.args = args
    self.result = result
    self.goresult = result.name
    self.interface = interface
    self.index = index
    self.structReturn = False

# builtin_types is a dictionary containing the predefined base types.
builtin_types = {
    '':        Type('void', 'void', True),
    'void':    Type('void', 'void', True),
    'bool':    Type('bool', 'int32', True),
    'int16':   Type('int16', 'int32', True),
    'uint16':  Type('uint16', 'int32', True),
    'int32':   Type('int32', 'int32', True),
    'uint32':  Type('uint32', 'int32', True),
    'int64':   Type('int64', 'int64', True),
    'uint64':  Type('uint64', 'int64', True),
    'uintptr': Type('uintptr', 'int32', True),
    'float32': Type('float32', 'float32', True),
    'float64': Type('float64', 'float64', True),
    }

class Processor:
  """Processor scans the header and template files and performs the template expansion"""

  # Const definitions scanned from the include files.
  consts = {}

  # Scanned configuration info.
  enums = {}
  types = builtin_types.copy()
  callbacks = []
  functions = []

  # typeof returns the type for a name.
  def typeof(self, ty):
    if ty.startswith('*'):
      return Type(ty, '*')
    if ty not in self.types:
      raise Exception("undefined type: " + ty)
    return self.types[ty]

  # parseArg parses an argument into a pair (name, type).
  def parseArg(self, arg, line):
    parts = arg.split()
    if len(parts) != 2:
      raise Exception("Argument should have the form 'name type': " + arg + " : " + line)
    return Arg(parts[0], self.typeof(parts[1]))

  # parseArgs splits the parameter list for a function.
  def parseArgs(self, args, line):
    if args == "":
      return []
    return [self.parseArg(arg.strip(), line) for arg in string.split(args, ',')]

  # framesize returns the total size for all arguments.
  def framesize(self, func):
    size = 0
    for arg in func.args:
      if arg.type.align > 0:
        size = align(size, arg.type.align)
      size += arg.type.size
    return size

  # updateStructReturn converts the function definition when the return value is
  # a struct that is larger than 8 bytes.  If so, the result is passed by
  # reference as the first arg.
  def updateStructReturn(self, func):
    if func.result.kind == 'struct':
      func.args.insert(0, Arg('return_struct', Type('*' + func.result.name, '*')))
      func.goresult = 'void'
      func.structReturn = True

  # findEnumType finds the type for a const.
  def findEnumType(self, c):
    for r, ty in self.enums.items():
      m = re.match(r, c)
      if m:
        return ty

  # scanIncludeFiles collect the constants from the include files.
  def scanIncludeFiles(self, files):
    for line in fileinput.input(files):
      m = re.search(r'(PP\w*)\s*=\s*([^,\n\r]+),?\s*$', line)
      if m:
        self.consts[m.group(1)] = m.group(2)

  # scanConfig reads the configuration from a set of files.
  def scanConfig(self, files):
    for line in fileinput.input(files):
      if line.startswith('//enum'):
        # //enum regex typename
        m = re.search(r'^//enum\s+(\w+)\s+(.*)$', line)
        if m == None:
          raise Exception('Malformed //enum: ' + line)
        self.enums[m.group(2)] = m.group(1)
        self.types[m.group(1)] = Type(m.group(1), 'int32')

      elif line.startswith('//type'):
        # //type name name
        m = re.search(r'^//type\s+(\w+)\s+([][,\w*]+)\s*(\w*)\s*$', line)
        if m == None:
          raise Exception('Malformed //type: ' + line)
        self.types[m.group(1)] = Type(m.group(1), m.group(2), m.group(3) <> '')

      elif line.startswith('//func'):
        # //func name(a1 t1, a2 t2, ..., aN tN) ty = x[i]
        m = re.search(r'^//func\s+(\w+)[(]([\w\s,*]*)[)]\s*([\w*]*)\s*=\s*(\w+)\s*\[\s*(\w+)\s*\]\s*$', line)
        if m == None:
          raise Exception("Malformed function: " + line)
        func = Func(m.group(1), self.parseArgs(m.group(2), line), self.typeof(m.group(3)), m.group(4), m.group(5))
        self.updateStructReturn(func)
        self.functions.append(func)

      elif line.startswith('//callback'):
        # //callback name(a1 t1, a2 t2, ..., aN tN) ty
        m = re.search(r'^//callback\s+(\w+)[(]([\w\s,*]*)[)]\s*([\w*]*)\s*$', line)
        if m == None:
          raise Exception("Malformed function: " + line)
        func = Func(m.group(1), self.parseArgs(m.group(2), line), self.typeof(m.group(3)), None, None)
        self.callbacks.append(func)

  # compile returns the dictionary used by the templates.
  def compile(self):
    enums = []
    for ty in self.enums.values():
      enums.append(ty)
    consts = []
    for c, v in self.consts.items():
      ty = self.findEnumType(c)
      if ty:
        consts.append((c, ty, v))
    return { 'callbacks': self.callbacks,
             'functions': self.functions,
             'consts': consts,
             'enums': enums,
             'types': self.types,
           }

  # printTemplates expands and prints all template files.
  def printTemplates(self, templates):
    state = self.compile()
    for file in templates:
      text = ''
      with open(file) as f:
        text = f.read()
      text = text.decode('utf-8')
      tmpl = Template(text)
      tmpl.globals['framesize'] = self.framesize
      tmpl.globals['max'] = max
      tmpl.globals['align'] = align
      text = tmpl.render(**state)
      print text.encode('utf-8')

# Preprocess the input files.
def main():
  args = parser.parse_args()
  print '//', string.join(sys.argv)
  print '// MACHINE GENERATED BY THE COMMAND ABOVE; DO NOT EDIT'
  print
  p = Processor()
  if args.include != None:
    for dir in args.include:
      p.scanIncludeFiles(glob.glob(dir + '/*.h'))
  p.scanConfig(args.template)
  p.printTemplates(args.template)

# This is the standard boilerplate that calls the main() function.
if __name__ == '__main__':
  main()
