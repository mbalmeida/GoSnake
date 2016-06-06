package main

import (
    "fmt"
    "github.com/nsf/termbox-go"
    "time"
    "errors"
    "math/rand"
)

type food_t struct {
    Y, X int
    good bool
    time time.Duration
}

type gameLogic_t struct {
    // x1, y1, x2, y1
    board_limits [4]int
    snake snake_t
    food []food_t
    lives, score, max_score int
    time time.Time
}

func draw(gamelogic *gameLogic_t) (int,error) {
    termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
    defer termbox.Flush()
    headpos := gamelogic.snake.Body[0]
    inbounds := headpos.Y > gamelogic.board_limits[1] && headpos.Y < gamelogic.board_limits[3]
    inbounds = inbounds && (headpos.X > gamelogic.board_limits[0] && headpos.X < gamelogic.board_limits[2])
    for i := 1; i < len(gamelogic.snake.Body); i++ {
        bodypos := gamelogic.snake.Body[i]
        hitbody := headpos.X == bodypos.X && headpos.Y == bodypos.Y
        if hitbody {
            draw_game_over(gamelogic)
            return 0, errors.New("TheEnd")
        }
    }
    if !inbounds {
        draw_game_over(gamelogic)
        return 0, errors.New("TheEnd")
    }
    count := 0
    for i := 0; i < len(gamelogic.food); i++ {
        if headpos.X == gamelogic.food[i].X && headpos.Y == gamelogic.food[i].Y {
            gamelogic.food[i].good = false
            count++
            gamelogic.score += 10
        }
    }
    draw_board(gamelogic)
    draw_snake(gamelogic.snake)
    draw_food(gamelogic)
    return count, nil
}

func draw_game_over(gamelogic *gameLogic_t) {
    w,h := termbox.Size()
    print_string_h(h/2, w/2, termbox.ColorDefault, termbox.ColorDefault, "Game Over")
}

func draw_snake(snake snake_t) {
    character := 'o'
    for _, c := range snake.Body {
        termbox.SetCell(c.X, c.Y, character, termbox.ColorYellow, termbox.ColorYellow)
    }
}

func draw_food(gamelogic *gameLogic_t) {
    character := 'X'
    for _, f := range gamelogic.food {
        if f.good {
            termbox.SetCell(f.X, f.Y, character, termbox.ColorGreen, termbox.ColorGreen)
        }
    }
}

func print_string_h(y, x int, fg, bg termbox.Attribute, str string) {
    s := fmt.Sprint(str)
    for _, r := range s {
        termbox.SetCell(x, y, r, fg, bg)
        x++
    }
}

func draw_board(gamelogic *gameLogic_t) {
    score := fmt.Sprintf("Score: %v", gamelogic.score)
    print_string_h(0, 0, termbox.ColorWhite | termbox.AttrBold, termbox.ColorDefault, score)
    maxscore := fmt.Sprintf("Max Score: %v", gamelogic.max_score)
    print_string_h(0, 20, termbox.ColorWhite | termbox.AttrBold, termbox.ColorDefault, maxscore)
    duration := fmt.Sprintf("Duration: %vs", int32(time.Now().Sub(gamelogic.time).Seconds()))
    print_string_h(0, 40, termbox.ColorWhite | termbox.AttrBold, termbox.ColorDefault, duration)

    width, height := termbox.Size()
    for j := 2; j < height; j++ {
        if j == 2 || j == height - 1 {
            for x := 0; x < width - 1; x++ {
                termbox.SetCell(x, j, '#', termbox.ColorRed, termbox.ColorRed)
            }
        }
        termbox.SetCell(0, j, '#', termbox.ColorRed, termbox.ColorRed)
        termbox.SetCell(width - 1, j, '#', termbox.ColorRed, termbox.ColorRed)
    }
}

type Pos struct {
    X, Y int
}

type snake_t struct {
    Body []Pos
}

const (
    DirectionUp = iota
    DirectionDown
    DirectionLeft
    DirectionRight
)

func genXY(gamelogic gameLogic_t) Pos {
    var x,y int
loop:
    for {
        x = rand.Int() % (gamelogic.board_limits[2] - 1)
        y = rand.Int() % (gamelogic.board_limits[3] - 1)
        inboundsx := x > 0
        inboundsy := y > 2
        if inboundsx && inboundsy {
            break loop
        }
    }
    return Pos{x,y}
}

func main() {
    err := termbox.Init()
    defer termbox.Close()

    if err != nil {
        panic(err)
    }

    width, height := termbox.Size()
    var gamelogic gameLogic_t
    gamelogic.board_limits[0] = 0
    gamelogic.board_limits[1] = 2
    w,h := termbox.Size()
    gamelogic.board_limits[2] = w-1
    gamelogic.board_limits[3] = h-1
    var snake snake_t
    game_direction := DirectionLeft

    snake.Body = append(snake.Body, Pos{width/2, height/2}, Pos{width/2 + 1, height/2})
    gamelogic.snake = snake
    gamelogic.lives = 3
    gamelogic.score = 10
    gamelogic.max_score = 0
    gamelogic.time = time.Now()

    termbox.SetInputMode(termbox.InputEsc)
    termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

    event_queue := make(chan termbox.Event)
    go func() {
       for {
           event_queue <- termbox.PollEvent()
       }
    }()

    _, errr := draw(&gamelogic)
    if errr != nil {
        time.Sleep(5 * time.Second)
        return
    }

    newfood := time.Now()
    var inc int
    var erra error
mainloop:
    for {
        select {
        case ev := <-event_queue:
            if ev.Type == termbox.EventKey {
                switch ev.Key {
                case termbox.KeyEsc:
                    break mainloop
                case termbox.KeyArrowUp:
                    game_direction = DirectionUp
                case termbox.KeyArrowDown:
                    game_direction = DirectionDown
                case termbox.KeyArrowLeft:
                    game_direction = DirectionLeft
                case termbox.KeyArrowRight:
                    game_direction = DirectionRight
                }
            }
        default:
            if inc > 0 {
                gamelogic.snake.Body = append(gamelogic.snake.Body,
                      Pos{gamelogic.snake.Body[len(gamelogic.snake.Body) - 1].X,
                          gamelogic.snake.Body[len(gamelogic.snake.Body) - 1].Y})
                inc = 0
            }
            for i := len(gamelogic.snake.Body) - 1; i > 0; i-- {
                gamelogic.snake.Body[i] = gamelogic.snake.Body[i-1]
            }
            switch game_direction {
            case DirectionUp:
                gamelogic.snake.Body[0].Y -= 1
            case DirectionDown:
                gamelogic.snake.Body[0].Y += 1
            case DirectionLeft:
                gamelogic.snake.Body[0].X -= 1
            case DirectionRight:
                gamelogic.snake.Body[0].X += 1
            }
        }
        t := time.Now()
        if (t.Sub(newfood).Seconds() >= 2) {
            pxy := genXY(gamelogic)
            gamelogic.food = append(gamelogic.food,  food_t{pxy.Y, pxy.X, true, time.Second * 5})
            newfood = time.Now()
        }
        inc, erra = draw(&gamelogic)
        if erra != nil {
            time.Sleep(1 * time.Second)
            break mainloop
        }
        time.Sleep(100 * time.Millisecond)
    }
}
