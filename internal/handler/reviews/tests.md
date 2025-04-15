go get github.com/stretchr/testify
go get github.com/golang/mock/gomock

mockgen -source=internal/domain/service/product-reviews.go -destination=internal/domain/service/reviews/product-review-mock.go -package=mocks

### TODO
- Поставить gomock глобально
