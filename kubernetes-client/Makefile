build:
	go build scheduler.go  utils.go mock.go
runa:
    # ----------------------------------------------
	# Remove all the existing end-to-end testing pods
	# ----------------------------------------------
	sh test/endtoendtest/python/clean_deploy.sh
	# ----------------------------------------------
	# ----------------------------------------------
	# Deploy all the end-to-end testing pods and they will wait for scheduler
	# ----------------------------------------------
	sh test/endtoendtest/python/deploy_images.sh
	# ----------------------------------------------
	# Start the athena scheduler
	# ----------------------------------------------
	go run scheduler.go utils.go mock.go 
	# ----------------------------------------------
rund:
    # ----------------------------------------------
    # Remove all the existing end-to-end testing pods
    # ----------------------------------------------
	sh test/endtoendtest/python/clean_deploy_default.sh
	# ----------------------------------------------
	# Deploy all the end-to-end testing pods and they will wait for scheduler
	# ----------------------------------------------
	sh test/endtoendtest/python/deploy_images_default.sh
show:
	kubectl -n athena get pods -o wide
runapp:
	sh test/endtoendtest/testapp.sh
clean:
	rm kubernetes-client main app

runroboa:
      # ----------------------------------------------
	sh test/endtoendtest/python/clean_deploy_roboimages_default.sh 
   	# ---------------------------------------------
	sh test/endtoendtest/python/deploy_roboimages.sh
 	# ---------------------------------------------
	go run scheduler.go utils.go mock.go
	# ---------------------------------------------

runrobod:
    # ---------------------------------------------
	sh test/endtoendtest/python/clean_deploy_roboimages.sh
	# ---------------------------------------------
	sh test/endtoendtest/python/deploy_roboimages_default.sh
	# ---------------------------------------------
