package net

func goLookupPort(network, service string) (port int, err error) {
    return 0, &AddrError{"unknown port", network + "/" + service}
}
