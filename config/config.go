package config

import "time"

// Secret 密钥
var Secret = "pointsmall"

// OneDayOfHours 时间
var OneDayOfHours = 60 * 60 * 24
var OneMinute = 60 * 1
var OneMonth = 60 * 60 * 24 * 30
var OneYear = 365 * 60 * 60 * 24
var ExpireTime = time.Hour * 48 // 设置Redis数据热度消散时间。

const ValidGood = 0   //商品状态：下单
const InvalidGood = 1 //商品状态：取消
const DateTime = "2006-01-02 15:04:05"
const KeyExist = 1
const KeyNotExist = 0
const ValidOrder = 0   //订单完成
const InvalidOrder = 1 //订单取消
const IsCoiled = 0     //是连续签到
const NotIsCoiled = 1  //不是连续签到
const IsDaily = 0      //是每日任务
const OnlyDay = 1      //一天的限时任务
const Attempts = 3     //操作数据库的最大尝试次数
const Finish = 0       //完成任务
const Rank = 0

const DefaultRedisValue = -1 //redis中key对应的预设值，防脏读
