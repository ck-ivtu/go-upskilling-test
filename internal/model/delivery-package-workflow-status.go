package model

type DeliveryPackageWorkflowStatus struct {
	RequestReceived bool
	RequestHandled  bool
}

func (s *DeliveryPackageWorkflowStatus) ShouldHandle() bool {
	return s.RequestReceived && !s.RequestHandled
}

func (s *DeliveryPackageWorkflowStatus) Handle() {
	s.RequestHandled = true
}
