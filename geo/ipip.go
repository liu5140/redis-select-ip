package geo

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"unsafe"
)

var offset int
var all, dataIndex []byte
var ipIndex = [65536]uint32{}

func Load(name string) error {
	var err error
	_, err = os.Stat(name)
	if err != nil {
		return err
	}

	all, err = ioutil.ReadFile(name)
	if err != nil {
		return err
	}
	offset = int(binary.BigEndian.Uint32(all[0:4]))

	for i := 0; i < 256; i++ {
		for j := 0; j < 256; j++ {
			k := i*256 + j
			ipIndex[k] = binary.LittleEndian.Uint32(all[(k+1)*4 : (k+1)*4+4])
		}
	}

	dataIndex = make([]byte, offset-4)
	copy(dataIndex, all[4:offset-4])

	return nil
}

func Find(s string) ([]byte, error) {
	if len(all) == 0 {
		return nil, fmt.Errorf("load ipdb file failed..")
	}
	ip := net.ParseIP(s)
	if ip == nil {
		return nil, fmt.Errorf("ip format error")
	}
	ip = ip.To4()
	var ipInt = binary.BigEndian.Uint32(ip)
	var prefix = int(ip[0])*256 + int(ip[1])
	var start = int(ipIndex[prefix])
	var maxValue = offset - 262144 - 4
	var b = make([]byte, 4)
	var indexOffset = -1
	var indexLength = -1

	for start = start*9 + 262144; start < maxValue; start += 9 {
		tmpInt := binary.BigEndian.Uint32(dataIndex[start : start+4])
		if tmpInt >= ipInt {
			b[1] = dataIndex[start+6]
			b[2] = dataIndex[start+5]
			b[3] = dataIndex[start+4]

			indexOffset = int(binary.BigEndian.Uint32(b))
			indexLength = 0xFF&int(dataIndex[start+7])<<8 + 0xFF&int(dataIndex[start+8])
			break
		}
	}

	if indexOffset == -1 || indexLength == -1 {
		return nil, fmt.Errorf("find failed.")
	}

	var area = make([]byte, indexLength)
	indexOffset = int(offset) + indexOffset - 262144
	copy(area, all[indexOffset:indexOffset+indexLength])

	return area, nil
}

func bytes2str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func Find2(s string) ([]string, error) {
	area, err := Find(s)
	if err != nil {
		return nil, err
	}

	return strings.Split(bytes2str(area), "\t"), nil
}

func Find3(s string) ([]string, error) {
	area, err := Find(s)
	if err != nil {
		return nil, err
	}

	return strings.Split(string(area), "\t"), nil
}

func FindLocation(s string) (*Location, error) {
	area, err := Find(s)
	if err != nil {
		return nil, err
	}
	var arr []string
	var loc *Location
	arr = strings.Split(string(area), "\t")
	loc = &Location{
		Country:   arr[0],
		Province:  arr[1],
		City:      arr[2],
		Org:       arr[3],
		Isp:       arr[4],
		Latitude:  arr[5],
		Longitude: arr[6],
		TimeZone:  arr[7],
		UTC:       arr[8],
		ChinaCode: arr[9],
		PhoneCode: arr[10],
		ISO2:      arr[11],
		Continent: arr[12],
	}

	return loc, nil
}

type Location struct {
	Country   string
	Province  string
	City      string
	Org       string
	Isp       string
	Latitude  string
	Longitude string
	TimeZone  string
	UTC       string
	ChinaCode string
	PhoneCode string
	ISO2      string
	Continent string
}
