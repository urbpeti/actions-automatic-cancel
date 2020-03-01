include .env

TEMPLATE = template.yaml
PACKAGED_TEMPLATE = packaged.yaml

.PHONY: configure
configure:
		aws s3api create-bucket \
			--bucket $(AWS_BUCKET_NAME) \
			--region $(AWS_REGION) \
			--create-bucket-configuration LocationConstraint=$(AWS_REGION)

.PHONY: clean
clean:
	@rm -rf dist
	@mkdir -p dist

.PHONY: build
build: clean
	@for dir in `ls handler`; do \
		GOOS=linux go build -o dist/handler/$$dir github.com/urbpeti/actions-automatic-cancel/handler/$$dir; \
	done
	rm -f packaged.yaml

.PHONY: run
run:
	sam local start-api

.PHONY: package
package: build
	sam package --template-file $(TEMPLATE) --s3-bucket $(AWS_BUCKET_NAME) --output-template-file $(PACKAGED_TEMPLATE)

	.PHONY: package
package: build
	sam package --template-file $(TEMPLATE) --s3-bucket $(AWS_BUCKET_NAME) --output-template-file $(PACKAGED_TEMPLATE)

.PHONY: deploy
deploy: package
	sam deploy --stack-name $(AWS_STACK_NAME) \
						 --template-file $(PACKAGED_TEMPLATE) \
						 --capabilities CAPABILITY_IAM \
						 --parameter-overrides \
						 	ApiSecret="$(WEBHOOK_SECRET)" \
						  GithubToken="$(GITHUB_TOKEN)" \
							GithubOrg="$(GITHUB_ORG)" \
							GithubRepo="$(GITHUB_REPO)"

.PHONY: teardown
teardown:
	aws cloudformation delete-stack --stack-name $(AWS_STACK_NAME)

.PHONY: test
test:
		go test ./... --cover