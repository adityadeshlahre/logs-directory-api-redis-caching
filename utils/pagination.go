package utils

import (
	"strconv"
)

type Pagination struct {
	Page  int
	Limit int
	Skip  int
}

func GetPagination(pageStr, limitStr string) (Pagination, error) {
	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 5
	}

	skip := (page - 1) * limit
	return Pagination{
		Page:  page,
		Limit: limit,
		Skip:  skip,
	}, nil
}
