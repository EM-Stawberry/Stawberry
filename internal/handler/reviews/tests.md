go get github.com/stretchr/testify
go get github.com/golang/mock/gomock

go install github.com/golang/mock/mockgen@latest
mockgen -source=internal/domain/service/reviews/product_reviews.go -destination=internal/domain/service/reviews/mocks/product_review_mock.go -package=mocks

go test ./internal/handler
go test ./...
