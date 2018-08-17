package interceptors // import "code.cloudfoundry.org/perm/pkg/api/rpc/interceptors"
import "path"

func parseFullMethod(fullMethod string) (string, string) {
	service := path.Dir(fullMethod)[1:]
	method := path.Base(fullMethod)

	return service, method
}
