# redis-select-ip
mysql+redis+go实现的本地ip库

1、把qqwry.txt数据导入到mysql中

建表语句

CREATE TABLE`geo_lite_city_blocks`  (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` datetime(0) NULL DEFAULT NULL,
  `updated_at` datetime(0) NULL DEFAULT NULL,
  `deleted_at` datetime(0) NULL DEFAULT NULL,
  `start_ip` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL,
  `end_ip` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL,
  `start_ip_num` int(255) NULL DEFAULT NULL,
  `end_ip_num` int(255) NULL DEFAULT NULL,
  `address` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL,
  `node_address` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL,
  `local_id` int(255) NULL DEFAULT NULL,
  `effective` int(255) NULL DEFAULT NULL,
  `country` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL,
  `province` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL,
  `city` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1808016 CHARACTER SET = utf8 COLLATE = utf8_general_ci ROW_FORMAT = Compact;


mysql 导入语句

set character_set_database=utf8;
load data local infile "C:/Users/user/go/src/geo/qqwry.txt" 
into table geo_lite_city_blocks(start_ip,end_ip,address,node_address);


删除不对的数据（不删会报错）
select *  FROM  geo_lite_city_blocks where  IFNULL(start_ip,0) =0;




2、直接运行即可（需要带数据库参数）
go run main.go  -dsn 'root:111111@tcp(127.0.0.1:3306)/dorama_core?charset=utf8&parseTime=True&loc=Local' -redisdsn '127.0.0.1:6379'

如何使用：

type GeoIP struct {
	ID          int64
	StartIP     string
	EndIP       string
	StartIPNum  int64
	EndIPNum    int64
	Address     string
	NodeAddress string
	LocalId     int64
	Country     string
	Province    string
	City        string
}


//通过ip 查询具体的省市区
func (service GeoipService) SearchCityByip(ip string) (genBlock GeoIP) {
	conn := credis.RedisClientManagerInstance().Client().Get()
	defer conn.Close()

	score := IPToScore(ip)
	if key, err := redis.Strings(conn.Do("zrangebyscore", "ipscore", score, "+inf", "LIMIT", "0", "1")); err != nil {
		return dto.GeoIP{}
	} else {
		if len(key) == 0 {
			return dto.GeoIP{}
		}
		//获取 城市 和 国家
		if valueGet, err := redis.Strings(conn.Do("hmget", "localip", key[0])); err == nil {
			json.Unmarshal([]byte(valueGet[0]), &genBlock)
			Log.Debugln("genBlock", genBlock)
			return genBlock
		} else {
			return GeoIP{}
		}
	}
}

func IPToScore(ipAddr string) int64 {
	score := 0
	if ipAddr == "::1" {
		ipAddr = "127.0.0.1"
	}
	//10.20.30.40的整数为 10*256*256*256 + 20*256*256 + 30*256 + 40 = 169090600
	ip := strings.Split(ipAddr, ".")
	ip1, _ := strconv.Atoi(ip[0])
	ip2, _ := strconv.Atoi(ip[1])
	ip3, _ := strconv.Atoi(ip[2])
	ip4, _ := strconv.Atoi(ip[3])
	score = ip1*256*256*256 + ip2*256*256 + ip3*256 + ip4
	return int64(score)
}
