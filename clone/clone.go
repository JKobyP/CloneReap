package clone

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

type Loc struct {
	Filename string
	Byte     uint
	End      uint
	Line     uint
}

func (l Loc) String() string {
	return fmt.Sprintf("%v:%v-%v", l.Filename, l.Byte, l.End)
}

type ClonePair struct {
	First  Loc
	Second Loc
}

func (c *ClonePair) String() string {
	return fmt.Sprintf("(1: %v; 2: %v)", c.First, c.Second)
}

func CloneParse(data string) ([]ClonePair, error) {
	files := make(map[int]string)
	pairs := make([]ClonePair, 0)
	readingSourceFiles := false
	readingClones := false
	for _, line := range strings.Split(data, "\n") {
		if strings.ContainsAny(line, "{}") {
			readingSourceFiles = false
			readingClones = false
			switch {
			case strings.Contains(line, "source_files"):
				readingSourceFiles = true
			case strings.Contains(line, "clone_pairs"):
				readingClones = true
			}
		} else if readingSourceFiles {
			words := strings.Split(line, "\t")
			fileid, err := strconv.Atoi(words[0])
			if err != nil {
				log.Println(err)
			}
			files[fileid] = strings.TrimSpace(words[1])
		} else if readingClones {
			words := strings.Split(line, "\t")
			if len(words) < 3 {
				return pairs, fmt.Errorf("Error parsing clones")
			}
			_, first, second := words[0], words[1], words[2]
			l1, err := parseLoc(first, files)
			if err != nil {
				return pairs, err
			}
			l2, err := parseLoc(second, files)
			if err != nil {
				return pairs, err
			}
			pairs = append(pairs, ClonePair{First: l1, Second: l2})
		}
	}
	return pairs, nil
}

func parseLoc(desc string, files map[int]string) (Loc, error) {
	str := strings.Split(desc, ".")
	fileId, err := strconv.Atoi(str[0])
	if err != nil {
		return Loc{}, err
	}
	interval := strings.Split(str[1], "-")
	bytenum, err := strconv.Atoi(interval[0])
	end, err2 := strconv.Atoi(interval[1])
	if err != nil {
		return Loc{}, err
	} else if err2 != nil {
		return Loc{}, err2
	}

	filename := files[fileId]
	return Loc{Filename: filename, Byte: uint(bytenum), End: uint(end)}, nil
}
