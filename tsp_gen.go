package main

import (
	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/llgcode/draw2d/draw2dkit"

	"image"
	"image/color"

	"math"
	"math/rand"
	"time"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"bufio"
	"os"
)

type Point struct {
	x	float64
	y	float64
	distances []float64
}

type Individual struct {
	way	[]int
	score float64
}
type IntSlice []int
type ByScore []Individual
func (a ByScore) Len() int           { return len(a) }
func (a ByScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByScore) Less(i, j int) bool { return a[i].score < a[j].score }



func (i * Individual) calculate_score(cities * []Point) {
	i.score = 0
	for idx, k := range i.way {
		var j int
		if idx+1 < len(i.way) {
			j = i.way[idx+1]
		} else {
			j = i.way[0]
		}
		i.score += (*cities)[k].distances[j]
	}
}

func (i * Individual) create_random_way(cities * []Point) {
	rand.Seed(time.Now().UTC().UnixNano())
	i.way = rand.Perm(len(*cities))
	i.calculate_score(cities);
}

func (i * Individual) mutate() {
	rand.Seed(time.Now().UTC().UnixNano())
	c1, c2 := rand.Intn(len(i.way) - 1), rand.Intn(len(i.way) - 1)
	i.way[c1], i.way[c2] = i.way[c2], i.way[c1]
}

func (slice IntSlice) is_in(value int) bool {
    for _, v := range slice {
        if (v == value) {
			return true
        }
    }
    return false
}

func (i Individual) crossover (partner * Individual) []Individual {
	pivot1 := rand.Intn(len(i.way) - 1)
	rand.Seed(time.Now().UTC().UnixNano())
	pivot2 := rand.Intn(len(i.way) - pivot1) + pivot1
	i1 := Individual{way: nil, score:0}
	i2 := Individual{way: nil, score:0}



	// for the first child
	for _, v := range i.way {
		if !IntSlice(partner.way[pivot1:pivot2+1]).is_in(v) {
			i1.way = append(i1.way, v)
		}
	}

	// dragons ahead!
	x := make([]int, 0)
	x = append(x, i1.way[:pivot1]...)
	x = append(x, partner.way[pivot1:pivot2+1]...)
	x = append(x, i1.way[pivot1:]...)
	i1.way = x

	// and for the second one
	for _, v := range partner.way {
		if !IntSlice(i.way[pivot1:pivot2+1]).is_in(v) {
			i2.way = append(i2.way, v)
		}
	}
	// dragons ahead pt2!
	x = make([]int, 0)
	x = append(x, i2.way[:pivot1]...)
	x = append(x, i.way[pivot1:pivot2+1]...)
	x = append(x, i2.way[pivot1:]...)
	i2.way = x

	return []Individual{i1, i2}
}

type Env struct {
	crossover_chance      float64
	mutation_chance       float64
	new_individual_factor float64
	choose_best_change    float64
	break_after           int

	max_generations       int
	population_size       int
	current_generation    int
	population_score      float64

	cities                []Point
	population            []Individual
}

func (e * Env) initialize() {
	e.calc_distances();
	e.population = make([]Individual, e.population_size)
	e.create_random_population()
}


func (e * Env) create_random_population() {
	for i := 0; i < e.population_size; i++ {
		e.population[i] = Individual{way: nil, score: 0}
		e.population[i].create_random_way(&(e.cities))
	}
	e.calc_score()
}

func (e * Env) calc_distances() {
	for i, v := range e.cities {
		e.cities[i].distances = make([]float64, len(e.cities))
		for k, v2 := range e.cities {
			dx, dy := v.x - v2.x, v.y - v2.y
			e.cities[i].distances[k] = math.Sqrt(dx * dx + dy * dy)
		}
	}
}

func (e * Env) calc_score() {
	e.population_score = 0
	for i, _ := range e.population {
		e.population[i].calculate_score(&e.cities)
		e.population_score += e.population[i].score
	}
	sort.Sort(ByScore(e.population))
}

func (e * Env) do_crossover() {
	rand.Seed(time.Now().UTC().UnixNano())
	if rand.Float64() < e.crossover_chance {
		crossover_count := int(e.new_individual_factor * float64(e.population_size))
		children := make([]Individual, 0)
		//fmt.Printf("Adding %d new children\n", crossover_count * 2)

		for i := 0; i < crossover_count; i++ {
			var p1 int
			var p2 int
			min := int(e.population_size / 3)
			// choose parents for crossover
			if rand.Float64() < e.choose_best_change {
				p1, p2 = rand.Intn(min), rand.Intn(min)
			} else {
				p1, p2 = rand.Intn(e.population_size - min) + min, rand.Intn(e.population_size - min) + min
			}
			children = append(children, e.population[p1].crossover(&e.population[p2])...)
		}
		for i := range children {
			children[i].calculate_score(&e.cities)
		}

		for _, v := range children {
			add_me_score := 0.5
			random_score := rand.Float64()
			for j, _ := range e.population {
				add_me_score += (e.population[len(e.population)-1-j].score / e.population_score)
				if add_me_score < 1 - random_score {
					e.population[len(e.population)-1-j] = v
					break
				}
			}
			e.calc_score()
		}
		fmt.Printf("Best score: %f\n", e.population[0].score)
	}
}


func (e * Env) do_mutation() {
	rand.Seed(time.Now().UTC().UnixNano())
	if rand.Float64() < e.mutation_chance {

		idx := rand.Intn(e.population_size / 3)
		p1, p2 := rand.Intn(len(e.population[idx].way)), rand.Intn(len(e.population[idx].way))
		e.population[idx].way[p1], e.population[idx].way[p2] = e.population[idx].way[p2], e.population[idx].way[p1]
		e.calc_score()
		//fmt.Printf("Mutation done\n")
	}
}

func (e * Env) run() {
	current_score := e.population_score
	no_changes := 0
	start := time.Now()
	for i := 0; i < e.max_generations; i++ {
		e.current_generation++
		fmt.Printf("Current generation: %d\n",  e.current_generation)
		e.do_crossover();
		e.do_mutation();

		if current_score == e.population_score {
			no_changes++
		} else {
			no_changes = 0
		}

		if no_changes == e.break_after {
			fmt.Printf("Stuck in local maximum for %d generations\n",  e.break_after)
			break
		}
		current_score = e.population_score
	}
	fmt.Printf("\n%s\n", time.Since(start))
	draw_way(e.population[0].way, e.cities)
}


func load_points() []Point {
	file, err := os.Open("./test.txt")
	if err != nil {
		fmt.Fprint(os.Stderr, "Cannot open points file")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	points := make([]Point, 0)
	for scanner.Scan() {
		p_data := strings.Split(scanner.Text(), " ")
		x, _ := strconv.ParseFloat(p_data[1], 64)
		y, _ := strconv.ParseFloat(p_data[2], 64)
		points = append(points, Point{
			x,
			y,
			nil,
		})
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprint(os.Stderr, "Points file reading error!")
	}
	return points
}


func draw_way(way []int, points []Point) {

	biggest_x, biggest_y := points[0].x, points[0].y
	for _, p := range points[1:] {
		if biggest_x < p.x {
			biggest_x = p.x
		}
		if biggest_y < p.y {
			biggest_y = p.y
		}
	}
	padding := 40.0
	resize_by := 10.0


	dest := image.NewRGBA(image.Rect(0, 0, int(biggest_x * resize_by + padding), int(biggest_y * resize_by + 2*padding)))
	gc := draw2dimg.NewGraphicContext(dest)


	gc.SetFillColor(color.RGBA{0x44, 0x44, 0x44, 0})
	gc.SetStrokeColor(color.RGBA{0x44, 0x44, 0x44, 0xff})
	gc.SetLineWidth(1)


	draw2dkit.Circle(gc, padding + points[way[0]].x * resize_by, padding + points[way[0]].y * resize_by, 3)
	gc.MoveTo(padding + points[way[0]].x * resize_by, padding + points[way[0]].y * resize_by)
	for _, v := range way[1:] {
		gc.LineTo(padding + points[v].x * resize_by, padding + points[v].y * resize_by)
		gc.MoveTo(padding + points[v].x * resize_by, padding + points[v].y * resize_by)
		gc.Close()
		draw2dkit.Circle(gc, padding + points[v].x * resize_by, padding + points[v].y * resize_by, 3)
	}
	gc.LineTo(padding + points[way[0]].x * resize_by , padding + points[way[0]].y * resize_by )
	gc.FillStroke()
	draw2dimg.SaveToPngFile("way.png", dest)
}

func main() {

	points := load_points()

	env := Env{
		max_generations: 	   3000,
		crossover_chance: 	   0.95,
		mutation_chance:  	   0.4,
		choose_best_change:	   0.95,
		current_generation:    0,
		new_individual_factor: 0.2,
		population_size:       100,
		break_after:		   100,
		population:            nil,
		cities:				   points,
	}


	env.initialize();
	env.run();
	fmt.Printf("%f\n", env.population[0].score)
	fmt.Printf("%d\n", env.max_generations)
	fmt.Printf("%f\n", env.crossover_chance)
	fmt.Printf("%f\n", env.mutation_chance)
	fmt.Printf("%f\n", env.choose_best_change)
	fmt.Printf("%f\n", env.new_individual_factor)

}
