package test

type TestService interface {
	Test(body any) any
}

type testService struct{}

func NewTestService() TestService {
	return &testService{}
}

func (s *testService) Test(body any) any {
	return body
}
