package httpadapter

import (
	"k8s.io/klog/v2"
	"net/url"
)

func filterQueryParams(queryParams string) (str string, urlValue url.Values, errors error) {
	m, err := url.ParseQuery(queryParams)
	if err != nil {
		errLog := "UnexpectedServerError failed to parse query parameter"
		klog.Error(errLog, err)
		return "", nil, err
	}

	var reserved = make(url.Values)
	// Separate parameters with SDK reserved prefix
	for k := range m {
		reserved.Set(k, m.Get(k))
		delete(m, k)
	}

	return m.Encode(), reserved, nil
}
