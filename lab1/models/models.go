package models

type SortRequest struct {
	Array []int  `json:"array"`
	Order string `json:"order"`
}

type SortResponse struct {
	OriginalArray []int `json:"original_array"`
	SortedArray   []int `json:"sorted_array"`
}

type ErrorMessage struct {
	Message string `json:"message"`
}
