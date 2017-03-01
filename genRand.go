import (
	."fmt"
	"math/rand"
)
	
// I stedet for å ha dette som en egen funksjon kan vi bare 
// kalle rand.Int63() direkte. (Hvis funksjonen uansett ikke skal 
// inneholde noe mer enn det den gjør nå)

func genRand() {
	// Må kun seedes én gang
	random := rand.Int63()
	Println(random)

}

