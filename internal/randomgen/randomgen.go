// Package randomgen provides functions to generate random data that can be used for tests.
package randomgen

import (
	"fmt"
	"math/rand"
	"time"
)

// PickFirstName picks a random American male or female first name.
func PickFirstName() string {
	randomIndex := rand.Intn(len(firstNames))
	return firstNames[randomIndex]
}

// PickLastName picks a random American last name.
func PickLastName() string {
	randomIndex := rand.Intn(len(lastNames))
	return lastNames[randomIndex]
}

// PickPhoneNumber generates a random 9-digit number and prefixes it with the specified conutry
// code.
func PickPhoneNumber(prefix string) string {
	first := rand.Intn(1000)
	middle := rand.Intn(1000)
	last := rand.Intn(1000)
	return fmt.Sprintf("%s %03d %03d %03d", prefix, first, middle, last)
}

// PickBirthDate selects a random date that is 18 to 78 years and a random number of days and
// months in the past.
func PickBirthDate() string {
	randYears := rand.Intn(60) + 18
	randMonths := rand.Intn(12)
	randDays := rand.Intn(31)
	birthday := time.Now().AddDate(-randYears, -randMonths, -randDays)
	return birthday.Format(time.RFC3339)
}

var firstNames = []string{
	// male names
	"Wade",
	"Dave",
	"Seth",
	"Ivan",
	"Riley",
	"Gilbert",
	"Jorge",
	"Dan",
	"Brian",
	"Roberto",
	"Ramon",
	"Miles",
	"Liam",
	"Nathaniel",
	"Ethan",
	"Lewis",
	"Milton",
	"Claude",
	"Joshua",
	"Glen",
	"Harvey",
	"Blake",
	"Noel",
	"Everett",
	"Romeo",
	"Sebastian",
	"Stefan",
	"Robin",
	"Clarence",
	"Sandy",
	"Ernest",
	"Samuel",
	"Benjamin",
	"Luka",
	"Fred",
	"Albert",
	"Greyson",
	"Terry",
	"Cedric",
	"Joe",
	"Paul",
	"George",
	"Bruce",
	"Christopher",
	"Stuart",
	"Orlando",
	"Keith",
	"Walter",
	"Marshall",
	"Shawn",

	// female names
	"Daisy",
	"Deborah",
	"Isabel",
	"Stella",
	"Debra",
	"Beverly",
	"Vera",
	"Angela",
	"Lucy",
	"Lauren",
	"Janet",
	"Loretta",
	"Tracey",
	"Beatrice",
	"Sabrina",
	"Melody",
	"Chrysta",
	"Christina",
	"Vicki",
	"Molly",
	"Alison",
	"Miranda",
	"Stephanie",
	"Leona",
	"Katrina",
	"Mila",
	"Teresa",
	"Gabriela",
	"Ashley",
	"Nicole",
	"Valentina",
	"Rose",
	"Juliana",
	"Alice",
	"Kathie",
	"Gloria",
	"Luna",
	"Phoebe",
	"Angelique",
	"Graciela",
	"Gemma",
	"Katelynn",
	"Danna",
	"Luisa",
	"Julie",
	"Olive",
	"Carolina",
	"Harmony",
	"Rachelle",
	"Kianna",
}

var lastNames = []string{
	"Salazar",
	"Combs",
	"Meadows",
	"Fischer",
	"Villegas",
	"Lucero",
	"Wilson",
	"Armstrong",
	"Irwin",
	"Dyer",
	"Dorsey",
	"Thompson",
	"Decker",
	"Cherry",
	"Jensen",
	"Gutierrez",
	"Brady",
	"Middleton",
	"Buck",
	"Bond",
	"Douglas",
	"Ellis",
	"Singleton",
	"Roman",
	"Randolph",
	"Hull",
	"Farmer",
	"Calhoun",
	"Powers",
	"Davidson",
	"Ray",
	"Manning",
	"Osborn",
	"Herman",
	"Forbes",
	"Horn",
	"Andrade",
	"Wade",
	"Alexander",
	"Travis",
	"Graves",
	"Chaney",
	"Guerra",
	"Rush",
	"Kane",
	"Harrington",
	"Keith",
	"Zimmerman",
	"House",
	"Haas",
	"Conrad",
	"Knox",
	"Horton",
	"Wilson",
	"Graves",
	"Shea",
	"Sherman",
	"Mathis",
	"Fisher",
	"Rowland",
	"Potter",
	"Brewer",
	"Gentry",
	"Ponce",
	"Eaton",
	"Rivera",
	"Blackburn",
	"Mercado",
	"Holden",
	"Vaughn",
	"Cherry",
	"Salinas",
	"Fuentes",
	"Kim",
	"Velasquez",
	"Giles",
	"Duran",
	"Mccall",
	"Rivas",
	"Riggs",
	"Bell",
	"Wilkinson",
	"Weiss",
	"Norris",
	"Ochoa",
	"Quinn",
	"Cruz",
	"Mitchell",
	"Ashley",
	"Love",
	"Pearson",
	"Logan",
	"Woodard",
	"Anthony",
	"Sims",
	"Farley",
	"Chaney",
	"Hebert",
	"Delgado",
	"Muller",
}