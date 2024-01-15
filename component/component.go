package component

import "github.com/bwmarrin/discordgo"

var ComponentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"flquantity": HandleModifyFoodQuantity,
	"fllist":     HandleUpdateList,
	"fldel":      HandleDeleteLog,
}
