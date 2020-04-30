#!/bin/sh

geo -dsn 'root:111111@tcp(127.0.0.1:3306)/dorama_core?charset=utf8&parseTime=True&loc=Local' -redisdsn '127.0.0.1:6379'
