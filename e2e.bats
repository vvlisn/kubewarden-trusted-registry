@test "accept when image is from a trusted registry" {
  # Run the policy with settings specifying trusted registries
  run kwctl run -r test_data/pod-trusted.json \
    --settings-json '{"trusted_registries": ["registry-dev.vestack.sbuxcf.net", "registry-stg.vestack.sbuxcf.net"]}' \
    policy.wasm

  # Print the output if any check fails
  echo "output = ${output}"

  # Check if the image is accepted from the trusted registry
  [ "$status" -eq 0 ]
  [ $(expr "$output" : '.*allowed.*true') -ne 0 ]
}

@test "reject when image is from an untrusted registry" {
  # Run the policy with settings specifying trusted registries
  run kwctl run -r test_data/pod-untrusted.json \
    --settings-json '{"trusted_registries": ["registry-dev.vestack.sbuxcf.net", "registry-stg.vestack.sbuxcf.net"]}' \
    policy.wasm

  # Print the output if any check fails
  echo "output = ${output}"

  # Check if the image is rejected from an untrusted registry (allowed: false)
  [ "$status" -eq 0 ]
  [ $(expr "$output" : '.*allowed.*false') -ne 0 ]
}


@test "accept when image is from one of multiple trusted registries" {
  # Run the policy with multiple trusted registries
  run kwctl run -r test_data/pod_multiple_images.json \
    --settings-json '{"trusted_registries": ["registry-dev.vestack.sbuxcf.net", "registry-stg.vestack.sbuxcf.net", "registry-prod.vestack.sbuxcf.net"]}' \
    policy.wasm

  # Print the output if any check fails
  echo "output = ${output}"

  # Check if the image from the trusted registry is accepted
  [ "$status" -eq 0 ]
  [ $(expr "$output" : '.*allowed.*true') -ne 0 ]
}

@test "reject when any container image is from an untrusted registry" {
  # Run the policy with a pod that has multiple containers, one of which uses an untrusted registry
  run kwctl run -r test_data/pod_multiple_containers_invalid.json \
    --settings-json '{"trusted_registries": ["registry-dev.vestack.sbuxcf.net", "registry-stg.vestack.sbuxcf.net", "registry-prod.vestack.sbuxcf.net"]}' \
    policy.wasm

  # Print the output if any check fails
  echo "output = ${output}"

  # Check if the request is rejected because one container has an untrusted image
  [ "$status" -eq 0 ]
  [ $(expr "$output" : '.*allowed.*false') -ne 0 ]
}








