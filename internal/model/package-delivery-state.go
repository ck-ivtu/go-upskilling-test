package model

type PackageDeliveryState string

const (
	PackageDeliveryInProgress PackageDeliveryState = "inProgress"
	PackageDeliveryConfirmed  PackageDeliveryState = "confirmed"
	PackageDeliverySaved      PackageDeliveryState = "confirmed"
	PackageDeliveryNotified   PackageDeliveryState = "confirmed"
	PackageDeliveryErrored    PackageDeliveryState = "errored"
)
