package procfs

import (
    "fmt"
    "strings"
    "strconv"

    "github.com/westonbelk/memcat/internal/util"
)



func parseMappingFields(entry []string) (Mapping, error) {
    if len(entry) < 5 {
        return Mapping{}, fmt.Errorf("unexpected number of fields in map entry: %d", len(entry))
    }

    if len(entry) > 6 {
        entry[5] = strings.Join(entry[5:], "")
    }

    addresses := strings.SplitN(entry[0], "-", 2)
    if len(addresses) != 2 {
        return Mapping{}, fmt.Errorf("unable to parse address in map entry")
    }

    startAddress, err := strconv.ParseUint(addresses[0], 16, 64)
    if err != nil {
        return Mapping{}, err
    }

    endAddress, err := strconv.ParseUint(addresses[1], 16, 64)
    if err != nil {
        return Mapping{}, err
    }

    if endAddress <= startAddress {
        return Mapping{}, fmt.Errorf("empty or impossible address range")
    }

    perms := entry[1]
    if len(perms) != 4 {
        return Mapping{}, fmt.Errorf("unexpected number of perm symbols: %d", len(perms))
    }
    permset := Permset{
        Read: perms[0] == 'r',
        Write: perms[1] == 'w',
        Execute: perms[2] == 'x',
        Shared: perms[3] == 's',
        Private: perms[3] == 'p',
    }

    offset, err := strconv.ParseUint(entry[2], 16, 64)
    if err != nil {
        return Mapping{}, err
    }

    devices := strings.SplitN(entry[3], ":", 2)
    if len(devices) != 2 {
        return Mapping{}, fmt.Errorf("unable to parse device in map entry")
    }

    deviceMajor, err := strconv.ParseUint(devices[0], 16, 64)
    if err != nil {
        return Mapping{}, err
    }

    deviceMinor, err := strconv.ParseUint(devices[1], 16, 64)
    if err != nil {
        return Mapping{}, err
    }

    inode, err := strconv.ParseUint(entry[4], 10, 64)
    if err != nil {
        return Mapping{}, err
    }

    pathname := ""
    if len(entry) == 6 {
        pathname = entry[5]
    }

    return Mapping{
        AddressStart: startAddress,
        AddressEnd: endAddress,
        Perms: permset,
        Offset: offset,
        Dev: Device{Major: deviceMajor, Minor: deviceMinor},
        Inode: inode,
        Pathname: pathname,
    }, nil
}


func ReadMap(pid Pid) ([]Mapping, error) {
    mapsLocation := fmt.Sprintf("%s/maps", pid.Dir())
    
    mapsFile, err := util.ReadLines(mapsLocation)
    if err != nil {
        return nil, err
    }
    maps := make([]Mapping, 0, len(mapsFile))

    for _, line := range mapsFile {
        mapFields := strings.Fields(line)
        parsedMapping, err := parseMappingFields(mapFields)

        if err != nil {
            return nil, err
        }
        maps = append(maps, parsedMapping)
    }

    return maps, nil
}


