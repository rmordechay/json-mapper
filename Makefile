run_tests:
	go test -v ./tests

push_tag:
	git tag v0.1.1
	git push --tags
