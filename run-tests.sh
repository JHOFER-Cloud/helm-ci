#!/bin/bash
# MIT License
#
# Copyright (c) 2025 Josef Hofer (JHOFER-Cloud)
#
# Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
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
package_coverages=()

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
    package_coverages+=("$pkg: $coverage%")
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
  echo "Package Coverage Summary:"
  for pkg_coverage in "${package_coverages[@]}"; do
    echo "  $pkg_coverage"
  done
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
