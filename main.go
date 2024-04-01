package main

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"regexp"
	"slices"
	"strings"
	"sync"
)

const (
	distanceBetweenWords = 2
	minLength            = 20
	maxLength            = 24
	wordsCount           = 4
	filepath             = "./data/linux_words"
)

var keyboardMap = map[rune][2]int{
	'q': {0, 0}, 'w': {0, 1}, 'e': {0, 2}, 'r': {0, 3}, 't': {0, 4}, 'y': {0, 5}, 'u': {0, 6}, 'i': {0, 7}, 'o': {0, 8}, 'p': {0, 9},
	'a': {1, 0}, 's': {1, 1}, 'd': {1, 2}, 'f': {1, 3}, 'g': {1, 4}, 'h': {1, 5}, 'j': {1, 6}, 'k': {1, 7}, 'l': {1, 8},
	'z': {2, 0}, 'x': {2, 1}, 'c': {2, 2}, 'v': {2, 3}, 'b': {2, 4}, 'n': {2, 5}, 'm': {2, 6},
}

type Word struct {
	word       []rune
	start, end rune
	distance   int
}

func (w *Word) Len() int {
	return len(w.word)
}

func (w *Word) String() string {
	return string(w.word)
}

type Passwords struct {
	mu                       sync.Mutex
	waitGroup                sync.WaitGroup
	minDist                  int
	passwordWordsWithMinDist [][wordsCount]string
	allPasswords             [][wordsCount]string
}

func main() {
	words, err := loadData(filepath)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(words), "words loaded")

	slices.SortFunc(words, func(i, j *Word) int { return (i.distance - i.Len()) - (j.distance - j.Len()) })

	// Init Defult Values
	left := 0
	wordsMinLen := make([]*Word, 0, len(words))
	passwords := &Passwords{
		minDist:                  math.MaxInt32,
		passwordWordsWithMinDist: make([][wordsCount]string, 0),
		allPasswords:             make([][wordsCount]string, 0),
	}

	// Init Params
	diffDistAndLen := words[0].distance - words[0].Len() + 1

	// Loop to find the best password
	// until at least one password satisfying the conditions is found
	for passwords.minDist == math.MaxInt32 {
		// Add words to make array with difference between distance and length <= diffDistAndLen
		for j := left; j < len(words); j++ {
			if words[j].distance-words[j].Len()-1 >= diffDistAndLen {
				left = j
				diffDistAndLen++
				break
			}
			wordsMinLen = append(wordsMinLen, words[j])
		}
		fmt.Printf("%d words with difference <= %d between distance and length\n", len(wordsMinLen), diffDistAndLen)
		fmt.Println("Distance between words <", distanceBetweenWords)

		wordGraph := makeWordGraph(distanceBetweenWords, wordsMinLen)

		findBestPasswordWithGraph(wordGraph, passwords)
	}

	fmt.Println("Min distance: ", passwords.minDist)
	fmt.Println("Total passwords With Min distance: ", len(passwords.passwordWordsWithMinDist))
	randomIndForMinDist := rand.Intn(len(passwords.passwordWordsWithMinDist))
	fmt.Println("Random password with min distance: ", strings.Join(passwords.passwordWordsWithMinDist[randomIndForMinDist][:], " "))
	fmt.Println("Total passwords: ", len(passwords.allPasswords))
	randomInd := rand.Intn(len(passwords.allPasswords))
	fmt.Println("Random password: ", strings.Join(passwords.allPasswords[randomInd][:], " "))

}

// Find best password with the given graph
//
// Given a word graph, find the shortest path
// to form a password with at least wordsCount words
// and is between minLength and maxLength words in length
//
// Return values:
// - Minimal distance between words in the best password
// - Passwords with minimal distance
// - All passwords satisfying the conditions
func findBestPasswordWithGraph(wordGraph map[*Word][]*Word, passwords *Passwords) {
	passwords.waitGroup.Add(len(wordGraph))

	for word1, relatedWords1 := range wordGraph {
		go func(word1 *Word, relatedWords1 []*Word) {
			defer passwords.waitGroup.Done()
			words := [wordsCount]*Word{word1}

			for _, word2 := range relatedWords1 {
				words[1] = word2
				if word1 == word2 {
					continue
				}
				proccessWordsLoops(2, wordGraph, words, passwords)
			}
		}(word1, relatedWords1)
	}

	passwords.waitGroup.Wait()
}

// Recursively iterates through all the possible combinations of words
// to find the best password with the given graph
func proccessWordsLoops(depth int, wordGraph map[*Word][]*Word, words [wordsCount]*Word, passwords *Passwords) {
	if depth == wordsCount {
		proccessPassword(words, passwords)
		return
	}
	for _, word := range wordGraph[words[depth-1]] {
		different := true
		for i := 0; i < depth; i++ {
			if words[i] == word {
				different = false
			}
		}
		if different {
			words[depth] = word
			proccessWordsLoops(depth+1, wordGraph, words, passwords)
		}
	}
}

// This function is called when a path of words is found
// It is responsible for adding the path to the list of found paths
func proccessPassword(words [wordsCount]*Word, passwords *Passwords) {
	length := words[0].Len()
	dist := words[0].distance
	wordsStrings := [wordsCount]string{words[0].String()}

	for i := 1; i < wordsCount; i++ {
		length += words[i].Len()
		wordsStrings[i] = words[i].String()
		dist += words[i].distance + distance(words[i-1].end, words[i].start)
	}

	if length >= minLength && length <= maxLength {
		passwords.mu.Lock()

		passwords.allPasswords = append(passwords.allPasswords, wordsStrings)

		if dist < passwords.minDist {
			passwords.minDist = dist
			passwords.passwordWordsWithMinDist = [][wordsCount]string{wordsStrings}
		} else if dist == passwords.minDist {
			passwords.passwordWordsWithMinDist = append(passwords.passwordWordsWithMinDist, wordsStrings)
		}

		passwords.mu.Unlock()
	}
}

// Creates a graph of words from the given words
// Edges between words are created if the distance between them is less than or equal to distanceBetweenWords
func makeWordGraph(distanceBetweenWords int, wordsMinLen []*Word) map[*Word][]*Word {
	wordGraph := make(map[*Word][]*Word, len(wordsMinLen))
	for i, word1 := range wordsMinLen {
		for j, word2 := range wordsMinLen {
			if i == j {
				continue
			}
			if distance(word1.end, word2.start) < distanceBetweenWords {
				wordGraph[word1] = append(wordGraph[word1], word2)
			}
		}
	}

	return wordGraph
}

// This function calculates the distance between two runes in a grid
// where each row represents a line of keyboard symbols.
func distance(a, b rune) int {
	ax, ay := keyboardMap[a][0], keyboardMap[a][1]
	bx, by := keyboardMap[b][0], keyboardMap[b][1]
	abs := func(x int) int {
		if x < 0 {
			return -x
		}
		return x
	}
	return abs(ax-bx) + abs(ay-by)
}

// Loads data from a file containing one word per line
// Each line should contain a single word consisting of only lowercase letters
// and no apostrophes
// Words should be separated by a newline
// The function returns a list of words and an error if any
func loadData(filepath string) ([]*Word, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	words := make([]*Word, 0)
	for scanner.Scan() {
		line := scanner.Text()
		re := regexp.MustCompile("^[a-z]+[^'s]$")
		if re.MatchString(strings.TrimSpace(line)) {
			distanceForWord := 0
			word := []rune(line)
			for i := 1; i < len(word); i++ {
				distanceForWord += distance(word[i-1], word[i])
			}
			words = append(words, &Word{
				word:     word,
				start:    word[0],
				end:      word[len(word)-1],
				distance: distanceForWord,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return words, nil
}
