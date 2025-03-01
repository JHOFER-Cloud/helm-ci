#!/bin/bash
set -e

# Go to project root directory
cd "$(dirname "$0")"

# Create output directory for coverage reports
mkdir -p coverage

# Run tests with coverage for each package
packages=(
  "./deploy/config"
  "./deploy/deployment"
  "./deploy/vault"
  "./deploy/utils"
)

total_coverage=0
package_count=0
failed_packages=()

go mod tidy

echo "================================="
echo "Running tests for helm-ci project"
echo "================================="

for pkg in "${packages[@]}"; do
  echo "Testing package: $pkg"
  pkg_name=$(basename "$pkg")

  # Run tests with coverage
  if ! go test -v -coverprofile="coverage/$pkg_name.out" "$pkg"; then
    failed_packages+=("$pkg")
    continue
  fi

  # Generate HTML coverage report
  go tool cover -html="coverage/$pkg_name.out" -o "coverage/$pkg_name.html"

  # Calculate coverage percentage
  coverage=$(go tool cover -func="coverage/$pkg_name.out" | grep total | awk '{print $3}' | tr -d '%')

  # Check if coverage is a valid number
  if [[ $coverage =~ ^[0-9]+(\.[0-9]+)?$ ]]; then
    total_coverage=$(echo "$total_coverage + $coverage" | bc)
    package_count=$((package_count + 1))
  fi

  echo "Package $pkg coverage: $coverage%"
  echo "------------------------------------"
done

# Calculate average coverage if any packages were tested
if [ "$package_count" -gt 0 ]; then
  # Combine all coverage files into one
  echo "mode: set" >coverage/all.out
  for pkg in "${packages[@]}"; do
    pkg_name=$(basename "$pkg")
    if [ -f "coverage/$pkg_name.out" ]; then
      grep -v "mode: set" "coverage/$pkg_name.out" >>coverage/all.out
    fi
  done

  # Generate combined HTML report
  go tool cover -html=coverage/all.out -o coverage/all.html

  # Calculate average and total coverage
  avg_coverage=$(echo "scale=2; $total_coverage / $package_count" | bc)
  total_coverage=$(go tool cover -func=coverage/all.out | grep total | awk '{print $3}')

  echo "================================="
  echo "Average package coverage: $avg_coverage%"
  echo "Total code coverage: $total_coverage"
  echo "Coverage report saved to: coverage/all.html"
else
  echo "No packages were tested successfully."
fi

# Print failed packages
if [ ${#failed_packages[@]} -gt 0 ]; then
  echo "================================="
  echo "The following packages had test failures:"
  for pkg in "${failed_packages[@]}"; do
    echo "  - $pkg"
  done
  exit 1
fi

echo "================================="
echo "All tests passed successfully!"
