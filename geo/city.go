package geo

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type City struct {
	Country       string
	Province      string
	City          string
	Org           string
	Isp           string
	Latitude      string
	Longitude     string
	TimeZone      string
	TimeZone2     string
	ChinaCode     string
	PhoneCode     string
	CountryISO2   string
	ContinentCode string
}

type CityDB struct {
	LastTime       time.Time
	ipIndex        [65536]uint32
	all, dataIndex []byte
	offset         int
	sync.RWMutex
}

func NewCityDB() *CityDB {
	return &CityDB{}
}

func (db *CityDB) Load(fileName string) error {
	db.Lock()
	defer db.Unlock()
	var err error
	info, err := os.Stat(fileName)
	if err != nil {
		return err
	}
	db.LastTime = info.ModTime()

	db.all, err = ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}

	str2 := string(binary.LittleEndian.Uint32(db.all))
	fmt.Println(str2)

	db.offset = int(binary.BigEndian.Uint32(db.all[0:4]))

	fmt.Println(db.offset)

	for i := 0; i < 256; i++ {
		for j := 0; j < 256; j++ {
			k := i*256 + j
			db.ipIndex[k] = binary.LittleEndian.Uint32(db.all[(k+1)*4 : (k+1)*4+4])
		}
	}

	db.dataIndex = make([]byte, db.offset-4)
	copy(db.dataIndex, db.all[4:db.offset-4])

	return nil
}

func (db *CityDB) _find(s string) ([]byte, error) {
	if len(db.all) == 0 {
		return nil, fmt.Errorf("error[%s]", "load ipdb file failed..")
	}
	ip := net.ParseIP(s)
	if ip == nil {
		return nil, fmt.Errorf("error[%s]", "ip format error")
	}

	ip = ip.To4()
	var ipInt = binary.BigEndian.Uint32(ip)
	var prefix = int(ip[0])*256 + int(ip[1])
	var start = int(ipIndex[prefix])
	var maxValue = db.offset - 262144 - 4
	var b = make([]byte, 4)
	var indexOffset = -1
	var indexLength = -1

	for start = start*9 + 262144; start < maxValue; start += 9 {
		tmpInt := binary.BigEndian.Uint32(db.dataIndex[start : start+4])
		if tmpInt >= ipInt {
			b[1] = db.dataIndex[start+6]
			b[2] = db.dataIndex[start+5]
			b[3] = db.dataIndex[start+4]

			indexOffset = int(binary.BigEndian.Uint32(b))
			indexLength = 0xFF&int(db.dataIndex[start+7])<<8 + 0xFF&int(db.dataIndex[start+8])
			break
		}
	}

	if indexOffset == -1 || indexLength == -1 {
		return nil, fmt.Errorf("error[%s]", "find failed.")
	}

	var area = make([]byte, indexLength)
	indexOffset = int(db.offset) + indexOffset - 262144
	copy(area, db.all[indexOffset:indexOffset+indexLength])

	return area, nil
}

func (db *CityDB) Find(s string) (City, error) {
	db.RLock()
	defer db.RUnlock()
	var err error
	var city City

	bs, err := db._find(s)
	if err != nil {
		return city, err
	}
	loc := strings.Split(string(bs), "\t")

	city.Country = loc[0]
	city.Province = loc[1]
	city.City = loc[2]
	city.Org = loc[3]
	city.Isp = loc[4]
	city.Latitude = loc[5]
	city.Longitude = loc[6]
	city.TimeZone = loc[7]
	city.TimeZone2 = loc[8]
	city.ChinaCode = loc[9]
	city.PhoneCode = loc[10]
	city.CountryISO2 = loc[11]
	city.ContinentCode = loc[12]

	return city, err
}
