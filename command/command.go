package command

import "github.com/bwmarrin/discordgo"

var (
	minCalorieIntake = 1.0
	maxItemCalories  = 5000.0

	minAverageDays = 2.0
	minQuantity    = 1.0

	CommandDefinitions = []*discordgo.ApplicationCommand{
		{
			Name:        "set",
			Description: "Set your daily calorie intake",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "calories",
					Description: "The amount of calories you need to intake daily",
					Required:    true,
					MinValue:    &minCalorieIntake,
					MaxValue:    maxItemCalories,
				},
			},
		},
		{
			Name:        "avg",
			Description: "Gives you an average of your calories consumed over X days",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "days",
					Description: "The amount of days to calculate the average for",
					MinValue:    &minAverageDays,
					MaxValue:    7,
					Required:    true,
				},
			},
		},
		{
			Name:        "rem",
			Description: "Gives your remaining calories for the day",
		},
		{
			Name:        "update",
			Description: "Update a log entry",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "logid",
					Description: "The ID of the log entry",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "fooditem",
					Description: "The name of the food product",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "calories",
					Description: "Calories consumed by eating this product",
					Required:    true,
					MinValue:    &minCalorieIntake,
					MaxValue:    maxItemCalories,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "quantity",
					Description: "The quantity consumed",
					Required:    false,
					MinValue:    &minQuantity,
				},
			},
		},
		{
			Name:        "list",
			Description: "List log entries for the current day or if the specified day if provided.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "date",
					Description: "The date to list",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "Which users list to view",
					Required:    false,
				},
			},
		},
		{
			Name:        "del",
			Description: "Delete a log entry",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "logid",
					Description: "The ID of the log entry",
					Required:    true,
				},
			},
		},
		{
			Name:        "add",
			Description: "Add an entry to your daily calories",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "fooditem",
					Description: "The name of the food product",
					Required:    true,
					MaxLength:   50,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "calories",
					Description: "The amount of calories consumed by eating this product",
					Required:    true,
					MinValue:    &minCalorieIntake,
					MaxValue:    maxItemCalories,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "quantity",
					Description: "The quantity consumed",
					Required:    false,
					MinValue:    &minQuantity,
				},
			},
		},
		{
			Name:        "conv",
			Description: "Figure out actual calories consumed when only given per X units",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionNumber,
					Name:        "units",
					Description: "The per X units (e.g grams or ml) given on the nutrition label",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionNumber,
					Name:        "calories",
					Description: "The amount of calories for the unit specified",
					Required:    true,
					MinValue:    &minCalorieIntake,
					MaxValue:    maxItemCalories,
				},
				{
					Type:        discordgo.ApplicationCommandOptionNumber,
					Name:        "weight",
					Description: "The actual weight in units of the packet",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "fooditem",
					Description: "The name of the food product",
					Required:    false,
				},
			},
		},
	}

	CommandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"set":    HandleSetCommand,
		"add":    HandleAddCommand,
		"update": HandleUpdateCommand,
		"del":    HandleDeleteCommand,
		"conv":   HandleConvCommand,
		"list":   HandleListCommand,
		"rem":    HandleRemCommand,
		"avg":    HandleAverageCommand,
	}
)
