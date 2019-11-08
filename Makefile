install:
	go install -v ./cmd/stacksearch

test:
	ginkgo -randomizeAllSpecs -randomizeSuites -p -r -race .
