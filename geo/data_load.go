package geo

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"liu5140/redis-select-ip/geo/model"
	"log"
	"strconv"
	"strings"

	"github.com/garyburd/redigo/redis"
)

type DataLoad struct {
}

type IPIPLocation struct {
	Country   string
	Province  string
	City      string
	District  string
	ChinaCode string
}

func (p *DataLoad) Load(sqd string, rediss string) {
	//1、连接数据库
	db := conn(sqd)
	defer db.Close()

	//2、连接redis
	conn, err := redis.Dial("tcp", rediss)
	if err != nil {
		fmt.Println("connect to redis err", err.Error())
		return
	}
	defer conn.Close()

	stmt, err := db.Prepare("update geo_lite_city_blocks set effective = 1 ,start_ip_num =?,end_ip_num=?,local_id = id,country=?,province=?,city=? where id = ?")
	h := 0
	country := make(map[string]model.Country)
	province := make(map[string]model.Provinces)
	cities := make(map[string]model.Cities)

	cityDb := NewCityDB()

	// dir, err := filepath.Abs(filepath.Dir(os.Args[0])) //返回绝对路径  filepath.Dir(os.Args[0])去除最后一个元素的路径
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// log.Println("dir", dir)

	err = cityDb.Load("./mydata4vipday2.datx")

	if err != nil {
		log.Fatal("err:======", err)
	}

	for {
		i := 0
		//查询表
		rows, err := db.Query("select id,local_Id,end_ip,end_ip_num,start_ip,start_ip_num,address,node_address,country,province,city from geo_lite_city_blocks where IFNULL(effective,0) =0 LIMIT 10000  ")
		defer rows.Close()

		if err != nil {
			panic(err)
		}

		var geolitecityblocks []model.GeoLiteCityBlocks
		for rows.Next() {
			var geo model.GeoLiteCityBlocks
			var id int64
			var localId sql.NullInt64
			var end_ip sql.NullString
			var end_ip_num sql.NullInt64
			var start_ip sql.NullString
			var start_ip_num sql.NullInt64
			var address sql.NullString
			var node_address sql.NullString
			var countrysql sql.NullString
			var provincesql sql.NullString
			var citysql sql.NullString
			err = rows.Scan(&id, &localId, &end_ip, &end_ip_num, &start_ip, &start_ip_num, &address, &node_address, &countrysql, &provincesql, &citysql)
			if err != nil {
				panic(err)
			}

			if localId.Valid {
				geo.LocalId = localId.Int64
			}
			if start_ip.Valid {
				geo.StartIP = start_ip.String
			}
			if end_ip.Valid {
				geo.EndIP = end_ip.String
			}
			if end_ip_num.Valid {
				geo.EndIPNum = end_ip_num.Int64
			}
			if address.Valid {
				geo.Address = address.String
			}
			if node_address.Valid {
				geo.NodeAddress = node_address.String
			}
			geolitecityblocks = append(geolitecityblocks, geo)

			startipnum := IPToScore(geo.StartIP)
			endipnum := IPToScore(geo.EndIP)
			var loc IPIPLocation
			if city, err := cityDb.Find(geo.StartIP); err == nil {
				loc.Country = city.Country
				loc.Province = city.Province
				loc.City = city.City
				loc.ChinaCode = city.ChinaCode
				//写入国家城市省份
				geo.Country = city.Country
				geo.Province = city.Province
				geo.City = city.City

				_, ok := country[city.Country]
				if !ok {
					countryNew := model.Country{Country: city.Country}
					country[city.Country] = countryNew
					//	log.Println("countryNew", countryNew)
				}
				_, ok1 := province[city.Province]
				if !ok1 {
					provinceNew := model.Provinces{Country: city.Country, Province: city.Province}
					province[city.Province] = provinceNew
					///	log.Println("provinceNew", provinceNew)
				}
				_, ok2 := cities[city.City]
				if !ok2 {
					citiesNew := model.Cities{Country: city.Country, Province: city.Province, City: city.City}
					cities[city.City] = citiesNew
					//	log.Println("citiesNew", citiesNew)
				}
			}

			geo.StartIPNum = startipnum
			geo.EndIPNum = endipnum
			_, err := stmt.Exec(startipnum, endipnum, loc.Country, loc.Province, loc.City, id)
			//	rows, err = db.Prepare("update geo_lite_city_blocks set effective = 1 ,start_ip_num =?,end_ip_num=? where id = ?", startipnum, endipnum, id)
			if err != nil {
				panic(err)
			}
			// _, err := res.RowsAffected()
			// if err != nil {
			// 	panic(err)
			// }
			//	fmt.Println(num)
			//defer rows.Close()

			i = i + 1
			if _, err := conn.Do("zadd", "ipscore_temp", geo.EndIPNum, id); err == nil {
				value, _ := json.Marshal(geo)
				_, err = conn.Do("hmset", "localip_temp", id, value)
				if err != nil {
					panic(err)
				}
			}
		}

		h = h + i
		log.Println("success", h)
		if i != 10000 {
			break
		}
	}
	// log.Println("country:%v", len(country))
	// log.Println("province:%v", len(province))
	// log.Println("cities:%v", len(cities))

	if h > 0 {
		//省市区
		// CountryInsert(db, country)
		// ProvinceInsert(db, province)
		// CityInsert(db, cities)
		//重刷
		conn.Do("MULTI")

		_, err = conn.Do("unlink", "ipscore", "localip")
		if err != nil {
			panic(err)
		}
		_, err = conn.Do("rename", "localip_temp", "localip")
		if err != nil {
			panic(err)
		}
		_, err = conn.Do("rename", "ipscore_temp", "ipscore")
		if err != nil {
			panic(err)
		}
		conn.Do("EXEC")
	}

	fmt.Println("success")
}
func CountryInsert(db *sql.DB, countrys map[string]model.Country) error {
	valueStrings := make([]string, 0, len(countrys))
	valueArgs := make([]interface{}, 0, len(countrys)*1)
	for _, v := range countrys {
		valueStrings = append(valueStrings, "(?)")
		valueArgs = append(valueArgs, v.Country)

	}
	insertsql := "INSERT INTO country (country) " +
		" VALUES %s"

	stmt := fmt.Sprintf(insertsql, strings.Join(valueStrings, ","))
	_, err := db.Exec(stmt, valueArgs...)
	if err != nil {
		log.Fatalf("err", err)
		panic(err)
	}
	return nil
}
func ProvinceInsert(db *sql.DB, provinces map[string]model.Provinces) error {
	valueStrings := make([]string, 0, len(provinces))
	valueArgs := make([]interface{}, 0, len(provinces)*2)
	for _, v := range provinces {
		valueStrings = append(valueStrings, "(?,?)")
		valueArgs = append(valueArgs, v.Country)
		valueArgs = append(valueArgs, v.Province)

	}
	insertsql := "INSERT INTO provinces (country,province) " +
		" VALUES %s"

	stmt := fmt.Sprintf(insertsql, strings.Join(valueStrings, ","))
	_, err := db.Exec(stmt, valueArgs...)
	if err != nil {
		log.Fatalf("err", err)
		panic(err)
	}
	return nil
}
func CityInsert(db *sql.DB, provinces map[string]model.Cities) error {
	valueStrings := make([]string, 0, len(provinces))
	valueArgs := make([]interface{}, 0, len(provinces)*3)
	for _, v := range provinces {
		valueStrings = append(valueStrings, "(?,?,?)")
		valueArgs = append(valueArgs, v.Country)
		valueArgs = append(valueArgs, v.Province)
		valueArgs = append(valueArgs, v.City)

	}
	insertsql := "INSERT INTO cities (country,province,city) " +
		" VALUES %s"

	stmt := fmt.Sprintf(insertsql, strings.Join(valueStrings, ","))
	_, err := db.Exec(stmt, valueArgs...)
	if err != nil {
		log.Fatalf("err", err)
		panic(err)
	}
	return nil
}

func IPToScore(ipAddr string) int64 {
	score := 0
	//10.20.30.40的整数为 10*256*256*256 + 20*256*256 + 30*256 + 40 = 169090600
	ip := strings.Split(ipAddr, ".")
	ip1, _ := strconv.Atoi(ip[0])
	ip2, _ := strconv.Atoi(ip[1])
	ip3, _ := strconv.Atoi(ip[2])
	ip4, _ := strconv.Atoi(ip[3])
	score = ip1*256*256*256 + ip2*256*256 + ip3*256 + ip4
	return int64(score)
}
