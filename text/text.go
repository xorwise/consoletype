package text

import (
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
)

type Alphabets map[string][]string

func GetAlphabets() Alphabets {
	var alphabets Alphabets

	jsonFile, err := os.Open("words.json")
	if err != nil {
		log.Fatal(err)
	}
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(byteValue, &alphabets)
	if err != nil {
		log.Fatal(err)
	}

	return alphabets
}

func GenerateText(lang string, length int) string {

	alphabets := GetAlphabets()

	var sb strings.Builder
	for i := 0; i < length; i++ {
		randIndex := rand.Intn(len(alphabets[lang]))
		sb.WriteString(alphabets[lang][randIndex])
		sb.WriteString(" ")
	}

	return sb.String()
}
