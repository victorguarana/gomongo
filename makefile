.PHONY: test

#=============================================================================

test:
	MONGO_VERSION=6 ginkgo -r
