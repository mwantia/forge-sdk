package random

import "math/rand"

var adjectives = []string{
	"admiring", "adoring", "agitated", "amazing", "angry", "awesome",
	"blissful", "bold", "boring", "brave", "busy", "calm",
	"charming", "clever", "cool", "compassionate", "confident", "cranky",
	"dazzling", "determined", "distracted", "dreamy", "eager", "elegant",
	"epic", "exciting", "fervent", "festive", "flamboyant", "focused",
	"friendly", "frosty", "funny", "gallant", "gifted", "gracious",
	"happy", "hardcore", "heuristic", "hopeful", "hungry", "infallible",
	"inspiring", "jolly", "kind", "laughing", "loving", "lucid",
	"magical", "musing", "mystifying", "naughty", "nervous", "nice",
	"nifty", "nostalgic", "objective", "optimistic", "peaceful", "pedantic",
	"pensive", "practical", "priceless", "quirky", "recursing", "relaxed",
	"reverent", "romantic", "sharp", "silly", "sleepy", "stoic",
	"stupefied", "suspicious", "sweet", "tender", "trusting", "unruffled",
	"upbeat", "vibrant", "vigilant", "vigorous", "wizardly", "wonderful",
	"xenodochial", "youthful", "zealous", "zen",
}

var nouns = []string{
	"albatross", "archimedes", "aristotle", "aryabhata", "babbage", "bardeen",
	"bohr", "booth", "bose", "burnell", "cannon", "carson",
	"cartwright", "cerf", "chandrasekhar", "chatelet", "chebyshev", "cohen",
	"curie", "darwin", "diffie", "dijkstra", "einstein", "elion",
	"euler", "faraday", "feynman", "fermat", "fermi", "franklin",
	"galileo", "gauss", "goldberg", "goldstine", "goodall", "hamilton",
	"hawking", "heisenberg", "hellman", "heyrovsky", "hofstadter", "hopper",
	"hypatia", "jackson", "jennings", "johnson", "joliot", "jones",
	"kalam", "kapitsa", "kare", "keldysh", "kepler", "khorana",
	"kilby", "knuth", "kowalevski", "lalande", "lamarr", "lamport",
	"leakey", "leavitt", "liskov", "lovelace", "mayer", "mccarthy",
	"mclean", "meitner", "mendel", "mendeleev", "mestorf", "mirzakhani",
	"montalcini", "moore", "morse", "napier", "neumann", "newton",
	"nightingale", "noether", "northcutt", "noyce", "panini", "pare",
	"pascal", "pasteur", "payne", "perlman", "pike", "planck",
	"ramanujan", "ride", "ritchie", "robinson", "rosalind", "sammet",
	"shaw", "shockley", "sinoussi", "stallman", "stonebraker", "swartz",
	"tesla", "thompson", "thorp", "torvalds", "turing", "varahamihira",
	"villani", "wescoff", "wilbur", "wiles", "wing", "wozniak",
	"wright", "wu", "yalow", "yonath",
}

// GenerateRandomName returns a random adjective_noun string.
func GenerateRandomName() string {
	adj := adjectives[rand.Intn(len(adjectives))]
	noun := nouns[rand.Intn(len(nouns))]
	return adj + "_" + noun
}
