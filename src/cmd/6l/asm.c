// Inferno utils/6l/asm.c
// http://code.google.com/p/inferno-os/source/browse/utils/6l/asm.c
//
//	Copyright © 1994-1999 Lucent Technologies Inc.  All rights reserved.
//	Portions Copyright © 1995-1997 C H Forsyth (forsyth@terzarima.net)
//	Portions Copyright © 1997-1999 Vita Nuova Limited
//	Portions Copyright © 2000-2007 Vita Nuova Holdings Limited (www.vitanuova.com)
//	Portions Copyright © 2004,2006 Bruce Ellis
//	Portions Copyright © 2005-2007 C H Forsyth (forsyth@terzarima.net)
//	Revisions Copyright © 2000-2007 Lucent Technologies Inc. and others
//	Portions Copyright © 2009 The Go Authors.  All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// Writing object files.

#include	"l.h"
#include	"../ld/lib.h"
#include	"../ld/elf.h"
#include	"../ld/dwarf.h"
#include	"../ld/macho.h"
#include	"../ld/pe.h"

#define PADDR(a)	((uint32)(a) & ~0x80000000)

char linuxdynld[] = "/lib64/ld-linux-x86-64.so.2";
char freebsddynld[] = "/libexec/ld-elf.so.1";
char openbsddynld[] = "/usr/libexec/ld.so";
char netbsddynld[] = "/libexec/ld.elf_so";

char	zeroes[32];

vlong
entryvalue(void)
{
	char *a;
	Sym *s;

	a = INITENTRY;
	if(*a >= '0' && *a <= '9')
		return atolwhex(a);
	s = lookup(a, 0);
	if(s->type == 0)
		return INITTEXT;
	if(s->type != STEXT)
		diag("entry not text: %s", s->name);
	return s->value;
}

vlong
datoff(vlong addr)
{
	if(addr >= segdata.vaddr)
		return addr - segdata.vaddr + segdata.fileoff;
	if(addr >= segtext.vaddr)
		return addr - segtext.vaddr + segtext.fileoff;
	diag("datoff %#llx", addr);
	return 0;
}

static int
needlib(char *name)
{
	char *p;
	Sym *s;

	if(*name == '\0')
		return 0;

	/* reuse hash code in symbol table */
	p = smprint(".elfload.%s", name);
	s = lookup(p, 0);
	free(p);
	if(s->type == 0) {
		s->type = 100;	// avoid SDATA, etc.
		return 1;
	}
	return 0;
}

int nelfsym = 1;

static void addpltsym(Sym*);
static void addgotsym(Sym*);

Sym *
lookuprel(void)
{
	return lookup(".rela", 0);
}

void
adddynrela(Sym *rela, Sym *s, Reloc *r)
{
	addaddrplus(rela, s, r->off);
	adduint64(rela, R_X86_64_RELATIVE);
	addaddrplus(rela, r->sym, r->add); // Addend
}

void
adddynrel(Sym *s, Reloc *r)
{
	Sym *targ, *rela, *got;
	
	targ = r->sym;
	cursym = s;

	switch(r->type) {
	default:
		if(r->type >= 256) {
			diag("unexpected relocation type %d", r->type);
			return;
		}
		break;

	// Handle relocations found in ELF object files.
	case 256 + R_X86_64_PC32:
		if(targ->dynimpname != nil && !(targ->cgoexport & CgoExportDynamic))
			diag("unexpected R_X86_64_PC32 relocation for dynamic symbol %s", targ->name);
		if(targ->type == 0 || targ->type == SXREF)
			diag("unknown symbol %s in pcrel", targ->name);
		r->type = D_PCREL;
		r->add += 4;
		return;
	
	case 256 + R_X86_64_PLT32:
		r->type = D_PCREL;
		r->add += 4;
		if(targ->dynimpname != nil && !(targ->cgoexport & CgoExportDynamic)) {
			addpltsym(targ);
			r->sym = lookup(".plt", 0);
			r->add += targ->plt;
		}
		return;
	
	case 256 + R_X86_64_GOTPCREL:
		if(targ->dynimpname == nil || (targ->cgoexport & CgoExportDynamic)) {
			// have symbol
			if(r->off >= 2 && s->p[r->off-2] == 0x8b) {
				// turn MOVQ of GOT entry into LEAQ of symbol itself
				s->p[r->off-2] = 0x8d;
				r->type = D_PCREL;
				r->add += 4;
				return;
			}
			// fall back to using GOT and hope for the best (CMOV*)
			// TODO: just needs relocation, no need to put in .dynsym
			targ->dynimpname = targ->name;
		}
		addgotsym(targ);
		r->type = D_PCREL;
		r->sym = lookup(".got", 0);
		r->add += 4;
		r->add += targ->got;
		return;
	
	case 256 + R_X86_64_64:
		if(targ->dynimpname != nil && !(targ->cgoexport & CgoExportDynamic))
			diag("unexpected R_X86_64_64 relocation for dynamic symbol %s", targ->name);
		r->type = D_ADDR;
		return;
	
	// Handle relocations found in Mach-O object files.
	case 512 + MACHO_X86_64_RELOC_UNSIGNED*2 + 0:
	case 512 + MACHO_X86_64_RELOC_SIGNED*2 + 0:
	case 512 + MACHO_X86_64_RELOC_BRANCH*2 + 0:
		// TODO: What is the difference between all these?
		r->type = D_ADDR;
		if(targ->dynimpname != nil && !(targ->cgoexport & CgoExportDynamic))
			diag("unexpected reloc for dynamic symbol %s", targ->name);
		return;

	case 512 + MACHO_X86_64_RELOC_BRANCH*2 + 1:
		if(targ->dynimpname != nil && !(targ->cgoexport & CgoExportDynamic)) {
			addpltsym(targ);
			r->sym = lookup(".plt", 0);
			r->add = targ->plt;
			r->type = D_PCREL;
			return;
		}
		// fall through
	case 512 + MACHO_X86_64_RELOC_UNSIGNED*2 + 1:
	case 512 + MACHO_X86_64_RELOC_SIGNED*2 + 1:
	case 512 + MACHO_X86_64_RELOC_SIGNED_1*2 + 1:
	case 512 + MACHO_X86_64_RELOC_SIGNED_2*2 + 1:
	case 512 + MACHO_X86_64_RELOC_SIGNED_4*2 + 1:
		r->type = D_PCREL;
		if(targ->dynimpname != nil && !(targ->cgoexport & CgoExportDynamic))
			diag("unexpected pc-relative reloc for dynamic symbol %s", targ->name);
		return;

	case 512 + MACHO_X86_64_RELOC_GOT_LOAD*2 + 1:
		if(targ->dynimpname == nil || (targ->cgoexport & CgoExportDynamic)) {
			// have symbol
			// turn MOVQ of GOT entry into LEAQ of symbol itself
			if(r->off < 2 || s->p[r->off-2] != 0x8b) {
				diag("unexpected GOT_LOAD reloc for non-dynamic symbol %s", targ->name);
				return;
			}
			s->p[r->off-2] = 0x8d;
			r->type = D_PCREL;
			return;
		}
		// fall through
	case 512 + MACHO_X86_64_RELOC_GOT*2 + 1:
		if(targ->dynimpname == nil || (targ->cgoexport & CgoExportDynamic))
			diag("unexpected GOT reloc for non-dynamic symbol %s", targ->name);
		addgotsym(targ);
		r->type = D_PCREL;
		r->sym = lookup(".got", 0);
		r->add += targ->got;
		return;
	}
	
	// Handle references to ELF symbols from our own object files.
	if(targ->dynimpname == nil || (targ->cgoexport & CgoExportDynamic))
		return;

	switch(r->type) {
	case D_PCREL:
		addpltsym(targ);
		r->sym = lookup(".plt", 0);
		r->add = targ->plt;
		return;
	
	case D_ADDR:
		if(s->type != SDATA)
			break;
		if(iself) {
			adddynsym(targ);
			rela = lookup(".rela", 0);
			addaddrplus(rela, s, r->off);
			if(r->siz == 8)
				adduint64(rela, ELF64_R_INFO(targ->dynid, R_X86_64_64));
			else
				adduint64(rela, ELF64_R_INFO(targ->dynid, R_X86_64_32));
			adduint64(rela, r->add);
			r->type = 256;	// ignore during relocsym
			return;
		}
		if(HEADTYPE == Hdarwin && s->size == PtrSize && r->off == 0) {
			// Mach-O relocations are a royal pain to lay out.
			// They use a compact stateful bytecode representation
			// that is too much bother to deal with.
			// Instead, interpret the C declaration
			//	void *_Cvar_stderr = &stderr;
			// as making _Cvar_stderr the name of a GOT entry
			// for stderr.  This is separate from the usual GOT entry,
			// just in case the C code assigns to the variable,
			// and of course it only works for single pointers,
			// but we only need to support cgo and that's all it needs.
			adddynsym(targ);
			got = lookup(".got", 0);
			s->type = got->type | SSUB;
			s->outer = got;
			s->sub = got->sub;
			got->sub = s;
			s->value = got->size;
			adduint64(got, 0);
			adduint32(lookup(".linkedit.got", 0), targ->dynid);
			r->type = 256;	// ignore during relocsym
			return;
		}
		break;
	}
	
	cursym = s;
	diag("unsupported relocation for dynamic symbol %s (type=%d stype=%d)", targ->name, r->type, targ->type);
}

int
elfreloc1(Reloc *r, vlong off, int32 elfsym, vlong add)
{
	VPUT(off);

	switch(r->type) {
	default:
		return -1;

	case D_ADDR:
		if(r->siz == 4)
			VPUT(R_X86_64_32 | (uint64)elfsym<<32);
		else if(r->siz == 8)
			VPUT(R_X86_64_64 | (uint64)elfsym<<32);
		else
			return -1;
		break;

	case D_PCREL:
		if(r->siz == 4)
			VPUT(R_X86_64_PC32 | (uint64)elfsym<<32);
		else
			return -1;
		add -= r->siz;
		break;
	}

	VPUT(add);
	return 0;
}

int
archreloc(Reloc *r, Sym *s, vlong *val)
{
	USED(r);
	USED(s);
	USED(val);
	return -1;
}

void
elfsetupplt(void)
{
	Sym *plt, *got;

	plt = lookup(".plt", 0);
	got = lookup(".got.plt", 0);
	if(plt->size == 0) {
		// pushq got+8(IP)
		adduint8(plt, 0xff);
		adduint8(plt, 0x35);
		addpcrelplus(plt, got, 8);
		
		// jmpq got+16(IP)
		adduint8(plt, 0xff);
		adduint8(plt, 0x25);
		addpcrelplus(plt, got, 16);
		
		// nopl 0(AX)
		adduint32(plt, 0x00401f0f);
		
		// assume got->size == 0 too
		addaddrplus(got, lookup(".dynamic", 0), 0);
		adduint64(got, 0);
		adduint64(got, 0);
	}
}

static void
addpltsym(Sym *s)
{
	if(s->plt >= 0)
		return;
	
	adddynsym(s);
	
	if(iself) {
		Sym *plt, *got, *rela;

		plt = lookup(".plt", 0);
		got = lookup(".got.plt", 0);
		rela = lookup(".rela.plt", 0);
		if(plt->size == 0)
			elfsetupplt();
		
		// jmpq *got+size(IP)
		adduint8(plt, 0xff);
		adduint8(plt, 0x25);
		addpcrelplus(plt, got, got->size);
	
		// add to got: pointer to current pos in plt
		addaddrplus(got, plt, plt->size);
		
		// pushq $x
		adduint8(plt, 0x68);
		adduint32(plt, (got->size-24-8)/8);
		
		// jmpq .plt
		adduint8(plt, 0xe9);
		adduint32(plt, -(plt->size+4));
		
		// rela
		addaddrplus(rela, got, got->size-8);
		adduint64(rela, ELF64_R_INFO(s->dynid, R_X86_64_JMP_SLOT));
		adduint64(rela, 0);
		
		s->plt = plt->size - 16;
	} else if(HEADTYPE == Hdarwin) {
		// To do lazy symbol lookup right, we're supposed
		// to tell the dynamic loader which library each 
		// symbol comes from and format the link info
		// section just so.  I'm too lazy (ha!) to do that
		// so for now we'll just use non-lazy pointers,
		// which don't need to be told which library to use.
		//
		// http://networkpx.blogspot.com/2009/09/about-lcdyldinfoonly-command.html
		// has details about what we're avoiding.

		Sym *plt;
		
		addgotsym(s);
		plt = lookup(".plt", 0);

		adduint32(lookup(".linkedit.plt", 0), s->dynid);

		// jmpq *got+size(IP)
		s->plt = plt->size;

		adduint8(plt, 0xff);
		adduint8(plt, 0x25);
		addpcrelplus(plt, lookup(".got", 0), s->got);
	} else {
		diag("addpltsym: unsupported binary format");
	}
}

static void
addgotsym(Sym *s)
{
	Sym *got, *rela;

	if(s->got >= 0)
		return;

	adddynsym(s);
	got = lookup(".got", 0);
	s->got = got->size;
	adduint64(got, 0);

	if(iself) {
		rela = lookup(".rela", 0);
		addaddrplus(rela, got, s->got);
		adduint64(rela, ELF64_R_INFO(s->dynid, R_X86_64_GLOB_DAT));
		adduint64(rela, 0);
	} else if(HEADTYPE == Hdarwin) {
		adduint32(lookup(".linkedit.got", 0), s->dynid);
	} else {
		diag("addgotsym: unsupported binary format");
	}
}

void
adddynsym(Sym *s)
{
	Sym *d, *str;
	int t;
	char *name;
	vlong off;

	if(s->dynid >= 0)
		return;

	if(s->dynimpname == nil)
		diag("adddynsym: no dynamic name for %s", s->name);

	if(iself) {
		s->dynid = nelfsym++;

		d = lookup(".dynsym", 0);

		name = s->dynimpname;
		if(name == nil)
			name = s->name;
		adduint32(d, addstring(lookup(".dynstr", 0), name));
		/* type */
		t = STB_GLOBAL << 4;
		if((s->cgoexport & CgoExportDynamic) && (s->type&SMASK) == STEXT)
			t |= STT_FUNC;
		else
			t |= STT_OBJECT;
		adduint8(d, t);
	
		/* reserved */
		adduint8(d, 0);
	
		/* section where symbol is defined */
		if(!(s->cgoexport & CgoExportDynamic) && s->dynimpname != nil)
			adduint16(d, SHN_UNDEF);
		else {
			switch(s->type) {
			default:
			case STEXT:
				t = 11;
				break;
			case SRODATA:
				t = 12;
				break;
			case SDATA:
				t = 13;
				break;
			case SBSS:
				t = 14;
				break;
			}
			adduint16(d, t);
		}
	
		/* value */
		if(s->type == SDYNIMPORT)
			adduint64(d, 0);
		else
			addaddr(d, s);
	
		/* size of object */
		adduint64(d, s->size);
	
		if(!(s->cgoexport & CgoExportDynamic) && s->dynimplib && needlib(s->dynimplib)) {
			elfwritedynent(lookup(".dynamic", 0), DT_NEEDED,
				addstring(lookup(".dynstr", 0), s->dynimplib));
		}
	} else if(HEADTYPE == Hdarwin) {
		// Mach-o symbol nlist64
		d = lookup(".dynsym", 0);
		name = s->dynimpname;
		if(name == nil)
			name = s->name;
		if(d->size == 0 && ndynexp > 0) { // pre-allocate for dynexps
			symgrow(d, ndynexp*16);
		}
		if(s->dynid <= -100) { // pre-allocated, see cmd/ld/go.c:^sortdynexp()
			s->dynid = -s->dynid-100;
			off = s->dynid*16;
		} else {
			off = d->size;
			s->dynid = off/16;
		}
		// darwin still puts _ prefixes on all C symbols
		str = lookup(".dynstr", 0);
		setuint32(d, off, str->size);
		off += 4;
		adduint8(str, '_');
		addstring(str, name);
		if(s->type == SDYNIMPORT) {
			setuint8(d, off, 0x01); // type - N_EXT - external symbol
			off++;
			setuint8(d, off, 0); // section
			off++;
		} else {
			setuint8(d, off, 0x0f);
			off++;
			switch(s->type) {
			default:
			case STEXT:
				setuint8(d, off, 1);
				break;
			case SDATA:
				setuint8(d, off, 2);
				break;
			case SBSS:
				setuint8(d, off, 4);
				break;
			}
			off++;
		}
		setuint16(d, off, 0); // desc
		off += 2;
		if(s->type == SDYNIMPORT)
			setuint64(d, off, 0); // value
		else
			setaddr(d, off, s);
		off += 8;
	} else if(HEADTYPE != Hwindows) {
		diag("adddynsym: unsupported binary format");
	}
}

void
adddynlib(char *lib)
{
	Sym *s;
	
	if(!needlib(lib))
		return;
	
	if(iself) {
		s = lookup(".dynstr", 0);
		if(s->size == 0)
			addstring(s, "");
		elfwritedynent(lookup(".dynamic", 0), DT_NEEDED, addstring(s, lib));
	} else if(HEADTYPE == Hdarwin) {
		machoadddynlib(lib);
	} else {
		diag("adddynlib: unsupported binary format");
	}
}

void
asmb(void)
{
	int32 magic;
	int i;
	vlong vl, symo, dwarfoff, machlink;
	Section *sect;
	Sym *sym;

	if(debug['v'])
		Bprint(&bso, "%5.2f asmb\n", cputime());
	Bflush(&bso);

	if(debug['v'])
		Bprint(&bso, "%5.2f codeblk\n", cputime());
	Bflush(&bso);

	if(iself)
		asmbelfsetup();

	sect = segtext.sect;
	cseek(sect->vaddr - segtext.vaddr + segtext.fileoff);
	codeblk(sect->vaddr, sect->len);

	/* output read-only data in text segment (rodata, gosymtab, pclntab, ...) */
	for(sect = sect->next; sect != nil; sect = sect->next) {
		cseek(sect->vaddr - segtext.vaddr + segtext.fileoff);
		datblk(sect->vaddr, sect->len);
	}

	if(debug['v'])
		Bprint(&bso, "%5.2f datblk\n", cputime());
	Bflush(&bso);

	cseek(segdata.fileoff);
	datblk(segdata.vaddr, segdata.filelen);

	machlink = 0;
	if(HEADTYPE == Hdarwin) {
		if(debug['v'])
			Bprint(&bso, "%5.2f dwarf\n", cputime());

		dwarfoff = rnd(HEADR+segtext.len, INITRND) + rnd(segdata.filelen, INITRND);
		cseek(dwarfoff);

		segdwarf.fileoff = cpos();
		dwarfemitdebugsections();
		segdwarf.filelen = cpos() - segdwarf.fileoff;

		machlink = domacholink();
	}

	switch(HEADTYPE) {
	default:
		diag("unknown header type %d", HEADTYPE);
	case Hplan9x32:
	case Hplan9x64:
	case Helf:
		break;
	case Hdarwin:
		debug['8'] = 1;	/* 64-bit addresses */
		break;
	case Hlinux:
	case Hfreebsd:
	case Hnetbsd:
	case Hopenbsd:
		debug['8'] = 1;	/* 64-bit addresses */
		break;
	case Hwindows:
		break;
	}

	symsize = 0;
	spsize = 0;
	lcsize = 0;
	symo = 0;
	if(!debug['s']) {
		if(debug['v'])
			Bprint(&bso, "%5.2f sym\n", cputime());
		Bflush(&bso);
		switch(HEADTYPE) {
		default:
		case Hplan9x64:
		case Helf:
			debug['s'] = 1;
			symo = HEADR+segtext.len+segdata.filelen;
			break;
		case Hdarwin:
			symo = rnd(HEADR+segtext.len, INITRND)+rnd(segdata.filelen, INITRND)+machlink;
			break;
		case Hlinux:
		case Hfreebsd:
		case Hnetbsd:
		case Hopenbsd:
			symo = rnd(HEADR+segtext.len, INITRND)+segdata.filelen;
			symo = rnd(symo, INITRND);
			break;
		case Hwindows:
			symo = rnd(HEADR+segtext.filelen, PEFILEALIGN)+segdata.filelen;
			symo = rnd(symo, PEFILEALIGN);
			break;
		}
		cseek(symo);
		switch(HEADTYPE) {
		default:
			if(iself) {
				cseek(symo);
				asmelfsym();
				cflush();
				cwrite(elfstrdat, elfstrsize);

				if(debug['v'])
				       Bprint(&bso, "%5.2f dwarf\n", cputime());

				dwarfemitdebugsections();
				
				if(isobj)
					elfemitreloc();
			}
			break;
		case Hplan9x64:
			asmplan9sym();
			cflush();

			sym = lookup("pclntab", 0);
			if(sym != nil) {
				lcsize = sym->np;
				for(i=0; i < lcsize; i++)
					cput(sym->p[i]);
				
				cflush();
			}
			break;
		case Hwindows:
			if(debug['v'])
			       Bprint(&bso, "%5.2f dwarf\n", cputime());

			dwarfemitdebugsections();
			break;
		}
	}

	if(debug['v'])
		Bprint(&bso, "%5.2f headr\n", cputime());
	Bflush(&bso);
	cseek(0L);
	switch(HEADTYPE) {
	default:
	case Hplan9x64:	/* plan9 */
		magic = 4*26*26+7;
		magic |= 0x00008000;		/* fat header */
		lputb(magic);			/* magic */
		lputb(segtext.filelen);			/* sizes */
		lputb(segdata.filelen);
		lputb(segdata.len - segdata.filelen);
		lputb(symsize);			/* nsyms */
		vl = entryvalue();
		lputb(PADDR(vl));		/* va of entry */
		lputb(spsize);			/* sp offsets */
		lputb(lcsize);			/* line offsets */
		vputb(vl);			/* va of entry */
		break;
	case Hplan9x32:	/* plan9 */
		magic = 4*26*26+7;
		lputb(magic);			/* magic */
		lputb(segtext.filelen);		/* sizes */
		lputb(segdata.filelen);
		lputb(segdata.len - segdata.filelen);
		lputb(symsize);			/* nsyms */
		lputb(entryvalue());		/* va of entry */
		lputb(spsize);			/* sp offsets */
		lputb(lcsize);			/* line offsets */
		break;
	case Hdarwin:
		asmbmacho();
		break;
	case Hlinux:
	case Hfreebsd:
	case Hnetbsd:
	case Hopenbsd:
		asmbelf(symo);
		break;
	case Hwindows:
		asmbpe();
		break;
	}
	cflush();
}

vlong
rnd(vlong v, vlong r)
{
	vlong c;

	if(r <= 0)
		return v;
	v += r - 1;
	c = v % r;
	if(c < 0)
		c += r;
	v -= c;
	return v;
}
