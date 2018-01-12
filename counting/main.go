package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/trustmaster/goflow"
)

type count struct {
	tag string
	count int
}

type splitter struct {
	flow.Component

	In <-chan string
	Out1, Out2 chan<- string
}

func (t *splitter) OnIn(s string) {
	t.Out1 <- s
	t.Out2 <- s
}

type wordCounter struct {
	flow.Component
	Sentence <-chan string
	Count chan<- *count
}

func (wc *wordCounter) OnSentence(sentence string) {
	wc.Count <- &count{"Words", len(strings.Split(sentence, " "))}
}

type letterCounter struct {
	flow.Component
	Sentence <-chan string
	Count chan<- *count

	re *regexp.Regexp
}

func (lc *letterCounter) OnSentence(sentence string) {
	lc.Count <- &count{"Letters", len(lc.re.FindAllString(sentence, -1))}
}

func (lc *letterCounter) Init() {
	lc.re = regexp.MustCompile("[a-zA-Z]")
}

type printer struct {
	flow.Component
	Line <-chan *count // inport
}

func (p *printer) OnLine(c *count) {
	fmt.Println(c.tag+":", c.count)
}

type counterNet struct {
	flow.Graph
}

func NewCounterNet() *counterNet {
	n := &counterNet{}
	n.InitGraphState()

	n.Add(&splitter{}, "splitter")
	n.Add(&wordCounter{}, "wordCounter")
	n.Add(&letterCounter{}, "letterCounter")
	n.Add(&printer{}, "printer")

	n.Connect("splitter", "Out1", "wordCounter", "Sentence")
	n.Connect("splitter", "Out2", "letterCounter", "Sentence")	
	n.Connect("wordCounter", "Count", "printer", "Line")	
	n.Connect("letterCounter", "Count", "printer", "Line")	

	n.MapInPort("In", "splitter", "In")
	return n 
}

func main ()  {
	net := NewCounterNet()
	in := make(chan string)
	net.SetInPort("In", in)
	
	flow.RunNet(net)

	in <- "Once upon a time"
	in <- "There was a bear who had big claws"
	in <- "It was scary for all who saw him."

	close(in)

	<-net.Wait()
}