package identity

import (
	"fmt"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var names = []string{
	"龙哥", "小李", "阿强", "大伟", "老王", "张三", "李四", "赵五",
	"陈总", "刘哥", "周老板", "吴师傅", "郑大侠", "孙师傅", "钱老板",
	"LuckyAce", "HighRoller", "BluffMaster", "RiverRat", "CardShark",
	"PokerFace", "AllInKing", "ChipLeader", "BadBeat", "FullHouse",
	"Kyaw", "Aung", "Myo", "Than", "Win", "Hla", "Khin", "Thida",
	"Zaw", "Min", "Soe", "Tun", "Naing", "Phyo", "Htay",
}

type Profile struct {
	ID      string
	Name    string
	Avatar  string
	VPIP    int
	PFR     int
	WinRate int
}

func GenerateProfile(index int) Profile {
	name := names[rand.Intn(len(names))]
	name = fmt.Sprintf("%s_%d", name, index)
	return Profile{
		ID:      fmt.Sprintf("bot_%d", index),
		Name:    name,
		Avatar:  fmt.Sprintf("https://api.dicebear.com/7.x/avataaars/svg?seed=%s", name),
		VPIP:    20 + rand.Intn(30),
		PFR:     10 + rand.Intn(25),
		WinRate: 40 + rand.Intn(20),
	}
}
