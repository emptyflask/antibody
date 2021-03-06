package antibodylib

import (
	"bufio"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/caarlos0/gohome"
	"github.com/getantibody/antibody/bundle"
	"golang.org/x/sync/errgroup"
)

// Antibody the main thing
type Antibody struct {
	r           io.Reader
	parallelism int
	Home        string
}

// New creates a new Antibody instance with the given parameters
func New(home string, r io.Reader, p int) *Antibody {
	return &Antibody{
		r:           r,
		parallelism: p,
		Home:        home,
	}
}

// Bundle processes all given lines and returns the shell content to execute
func (a *Antibody) Bundle() (result string, err error) {
	var g errgroup.Group
	var lock sync.Mutex
	var shs indexedLines
	var idx int
	sem := make(chan bool, a.parallelism)
	scanner := bufio.NewScanner(a.r)
	for scanner.Scan() {
		l := scanner.Text()
		index := idx
		idx++
		sem <- true
		g.Go(func() error {
			defer func() {
				<-sem
			}()
			l = strings.TrimSpace(l)
			if l == "" || l[0] == '#' {
				return nil
			}
			s, berr := bundle.New(a.Home, l).Get()
			lock.Lock()
			shs = append(shs, indexedLine{idx: index, line: s})
			lock.Unlock()
			return berr
		})
	}
	if err = scanner.Err(); err != nil {
		return
	}
	err = g.Wait()
	return shs.String(), err
}

// Home finds the right home folder to use
func Home() string {
	home := os.Getenv("ANTIBODY_HOME")
	if home == "" {
		home = gohome.Cache("antibody")
	}
	return home
}
