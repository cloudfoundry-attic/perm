package logging

type SecurityLoggerSignature string
type SecurityLoggerName string
type VendorName string
type ProductName string
type Version string
type CEFLogger struct {
	vendorName  VendorName
	productName ProductName
	version     Version
}

func NewCEFLogger(vendor VendorName, product ProductName, version Version) *CEFLogger {
	return &CEFLogger{vendor, product, version}
}

func (l *CEFLogger) Log(signature SecurityLoggerSignature, name SecurityLoggerName) {

}
