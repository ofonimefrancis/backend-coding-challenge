package movies

type Rating string

const (
	RatingG          Rating = "G"
	RatingPG         Rating = "PG"
	RatingPG13       Rating = "PG13"
	RatingRestricted Rating = "Restricted"
	RatingNC17       Rating = "NC17"
)

func AllowedRatings() []string {
	return []string{
		string(RatingG),
		string(RatingPG),
		string(RatingPG13),
		string(RatingRestricted),
		string(RatingNC17),
	}
}

func IsValid(rating string) bool {
	for _, allowed := range AllowedRatings() {
		if rating == allowed {
			return true
		}
	}
	return false
}
