#!/bin/sh

geo -dsn 'root:111111@tcp(127.0.0.1:3306)/dorama_core?charset=utf8&parseTime=True&loc=Local' -redisdsn '127.0.0.1:6379'

geo -dsn 'dbrentuser:kAaSWE$%@tcp(tydb.cdc5ggu7i40t.ap-northeast-1.rds.amazonaws.com:3306)/dorama_core?charset=utf8&parseTime=True&loc=Local' -redisdsn 'tyredis.eb0ynq.ng.0001.apne1.cache.amazonaws.com:6379'