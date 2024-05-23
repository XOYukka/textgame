package main

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"
)

var player Player

var locations = make(map[string]*Location)

// CommandHandler Принимает всю строку команды, вводимую пользователем 'идти комната'
type CommandHandler func([]string) (string, error)

var commandFuncResponse = map[string]CommandHandler{
	"идти":        walk,
	"взять":       take,
	"осмотреться": look,
}

type GameError struct {
	Errors []string
}

func (ge GameError) Error() string {
	return strings.Join(ge.Errors, " ")
}

type DescHandler func() string
type Location struct {
	Name          string
	Discriptions  map[string]DescHandler
	NearLocations []*Location
	Items         []*Item
}

type Player struct {
	Name      string
	Location  *Location
	Inventory []*Item
}

type Item struct {
	Name string
}

func (i *Item) String() string {
	return i.Name
}

func printItems(items []*Item) string {
	var result string

	for _, i := range items {
		result += i.Name
		result += ", "
	}
	return result[:len(result)-2]

}

func (l *Location) interconnectTo(toConnect ...*Location) {
	if len(toConnect) == 0 {
		panic("Incorrect call\t(l *Location) connectTo(toConnect ...*Location) ")
	}
	for _, loc := range toConnect {
		if !slices.Contains(loc.NearLocations, l) {
			loc.NearLocations = append(loc.NearLocations, l)
		}
		if !slices.Contains(l.NearLocations, loc) {
			l.NearLocations = append(l.NearLocations, loc)
		}
	}
}

func initLocation() {
	var (
		kitchen,
		room,
		hallway,
		street *Location
	)

	kitchen = &Location{
		Name: "кухня",
		Discriptions: map[string]DescHandler{
			"осмотреться": func() string {
				return "ты находишься на кухне, надо собрать рюкзак и идти в универ. можно пройти - коридор"
			},
			"идти": func() string {
				return "кухня, ничего интересного. можно пройти - коридор"
			},
		},
		NearLocations: []*Location{},
		Items:         []*Item{},
	}

	locations[kitchen.Name] = kitchen

	room = &Location{
		Name: "комната",
		Discriptions: map[string]DescHandler{
			"осмотреться": func() string {
				uniqItems := player.uniqueIntems()
				if len(uniqItems) == 0 {
					return "пустая комната. можно пройти - коридор"
				}

				return "на столе: " + printItems(player.uniqueIntems()) + ". можно пройти - коридор"
			},
			"идти": func() string { return "ты в своей комнате. можно пройти - коридор" },
		},
		NearLocations: []*Location{},
		Items: []*Item{{
			Name: "ключи",
		}, {
			Name: "конспекты",
		}, {
			Name: "рюкзак",
		}},
	}

	locations[room.Name] = room

	hallway = &Location{
		Name: "коридор",
		Discriptions: map[string]DescHandler{
			"осмотреться": func() string {
				return "ничего интересного. можно пройти - кухня, комната, улица"
			},
			"идти": func() string {
				return "ничего интересного. можно пройти - кухня, комната, улица"
			},
		},
		NearLocations: []*Location{},
		Items:         []*Item{},
	}

	locations[hallway.Name] = hallway

	street = &Location{
		Name: "улица",
		Discriptions: map[string]DescHandler{
			"осмотреться": func() string { return "на улице весна. можно пройти - домой" },
			"идти":        func() string { return "на улице весна. можно пройти - домой" },
		},
		NearLocations: []*Location{},
		Items:         []*Item{},
	}

	locations[street.Name] = street

	kitchen.interconnectTo(hallway)
	hallway.interconnectTo(kitchen, room, street)
	room.interconnectTo(hallway)
	// street.interconnectTo(hallway)

}

func walk(args []string) (string, error) {

	if len(args) != 2 {
		return "", GameError{
			Errors: []string{"Неверная команда! Необходимо -> идти 'место'"},
		}
	}

	if _, ok := locations[args[1]]; !ok {
		return "", GameError{
			Errors: []string{"нет пути в", args[1]},
		}
	}

	canWalk := false

	for _, nearLocation := range player.Location.NearLocations {
		if nearLocation.Name == args[1] {
			canWalk = true
			break
		}
	}

	if !canWalk {
		return "", GameError{
			Errors: []string{"нет пути в", args[1]},
		}
	}

	player.Location = locations[args[1]]

	return locations[args[1]].Discriptions[args[0]](), nil
}

func (p *Player) uniqueIntems() []*Item {

	var uniqueIntems []*Item

	for _, locItem := range player.Location.Items {
		foundInInventory := false
		for _, takenItem := range player.Inventory {
			if locItem == takenItem {
				foundInInventory = true
			}
		}
		if foundInInventory {
			continue
		}
		uniqueIntems = append(uniqueIntems, locItem)
	}

	return uniqueIntems
}

func take(args []string) (string, error) {

	foundItem := false

	for _, item := range player.uniqueIntems() {
		if item.Name == args[1] {
			player.Inventory = append(player.Inventory, item)
			foundItem = true
			break
		}
	}

	if !foundItem {
		return "", GameError{
			Errors: []string{"нет такого"},
		}
	}

	return "предмет добавлен в инвентарь: " + args[1], nil
}

func look(args []string) (string, error) {

	if len(args) > 1 {
		return "", GameError{
			Errors: []string{"Неверная команда! Необходимо -> осмотреться"},
		}
	}

	return player.Location.Discriptions[args[0]](), nil
}

func main() {
	/*
	  в этой функции можно ничего не писать,
	  но тогда у вас не будет работать через go run main.go
	  очень круто будет сделать построчный ввод команд тут, хотя это и не требуется по заданию
	*/
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("\twrite command\n\t\tor\n\texit to quit\n")
	initGame()
	for scanner.Scan() {
		req := scanner.Text()
		if req == "выход" {
			os.Exit(0)
		}
		resp := handleCommand(req)
		fmt.Println(">\t", resp)
	}
}

func initGame() {
	/*
		эта функция инициализирует игровой мир - все локации
		если что-то было - оно корректно перезатирается
	*/
	initLocation()

	if len(commandFuncResponse) == 0 {
		commandFuncResponse["осмотреться"] = look
		commandFuncResponse["идти"] = walk
		commandFuncResponse["взять"] = take
	}

	player = Player{
		Name:      "Player",
		Location:  locations["кухня"],
		Inventory: []*Item{},
	}

}

func handleCommand(command string) string {
	/*
		данная функция принимает команду от "пользователя"
		и наверняка вызывает какой-то другой метод или функцию у "мира" - списка комнат
	*/
	data := strings.Fields(command)
	if _, ok := commandFuncResponse[data[0]]; !ok {
		return "неизвестная команда"
	}

	response, err := commandFuncResponse[data[0]](data)
	if err != nil {
		return err.Error()
	}

	return response
}
