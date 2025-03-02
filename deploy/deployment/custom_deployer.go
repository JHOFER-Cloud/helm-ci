// Copyright 2025 Josef Hofer (JHOFER-Cloud)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// FIX: metadata.namespace in manifest file has to be set/replaced
package deployment

import (
	"bytes"
	"helm-ci/deploy/utils"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// CustomDeployer implements custom Kubernetes manifest deployments
type CustomDeployer struct {
	Common
}

// Deploy implements the custom deployment
func (d *CustomDeployer) Deploy() error {
	stageManifests, err := filepath.Glob(filepath.Join(d.Config.ValuesPath, d.Config.Stage, "*.y*ml"))
	if err != nil {
		return utils.NewError("failed to glob stage manifests: %w", err)
	}

	commonManifests, err := filepath.Glob(filepath.Join(d.Config.ValuesPath, "common", "*.y*ml"))
	if err != nil {
		return utils.NewError("failed to glob common manifests: %w", err)
	}
	manifests := append(stageManifests, commonManifests...)

	// Process manifests with Vault templating and update namespaces
	processedManifests := make([]string, 0, len(manifests))
	for _, manifest := range manifests {
		// First process with Vault templating
		processedFile, err := d.ProcessValuesFileWithVault(manifest)
		if err != nil {
			return err
		}
		if processedFile != manifest {
			defer os.Remove(processedFile)
		}

		// Then update namespaces in the processed file
		finalFile, err := d.updateNamespaces(processedFile)
		if err != nil {
			return err
		}
		if finalFile != processedFile && finalFile != manifest {
			defer os.Remove(finalFile)
		}

		processedManifests = append(processedManifests, finalFile)
	}

	// Check if namespace exists, create if it doesn't
	cmd := d.Cmd.Command("kubectl", "get", "namespace", d.Config.Namespace)
	if err := d.Cmd.Run(cmd); err != nil {
		utils.Green("Namespace %s does not exist, creating it...", d.Config.Namespace)
		cmd = d.Cmd.Command("kubectl", "create", "namespace", d.Config.Namespace)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := d.Cmd.Run(cmd); err != nil {
			return utils.NewError("failed to create namespace %s: %v", d.Config.Namespace, err)
		}
	}

	// Show diff first
	utils.Green("Showing differences:")
	if err := d.GetDiff(processedManifests, false); err != nil {
		return err
	}

	// Check if we should proceed
	if !utils.ConfirmDeployment(d.Config.DEBUG) {
		return utils.NewError("Deployment cancelled by user")
	}

	// Proceed with actual deployment
	for _, manifest := range processedManifests {
		cmd := d.Cmd.Command("kubectl", "apply", "-f", manifest, "-n", d.Config.Namespace)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := d.Cmd.Run(cmd); err != nil {
			return utils.NewError("failed to apply manifest %s: %v", manifest, err)
		}
	}

	return nil
}

// updateNamespaces processes YAML manifest files and ensures that
// metadata.namespace is set to the correct namespace for each resource
func (d *CustomDeployer) updateNamespaces(manifestFile string) (string, error) {
	// Read the manifest file
	content, err := os.ReadFile(manifestFile)
	if err != nil {
		return "", utils.NewError("failed to read manifest file %s: %v", manifestFile, err)
	}

	// Split the file into YAML documents
	yamlDocuments := bytes.Split(content, []byte("---"))
	if len(yamlDocuments) == 0 {
		return manifestFile, nil // Empty file
	}

	// Track if any document was updated
	updated := false
	updatedDocuments := make([][]byte, 0, len(yamlDocuments))

	// Process each document separately
	for _, docBytes := range yamlDocuments {
		// Skip empty documents
		if len(bytes.TrimSpace(docBytes)) == 0 {
			updatedDocuments = append(updatedDocuments, docBytes)
			continue
		}

		// Try to parse the document
		var docNode yaml.Node
		err := yaml.Unmarshal(docBytes, &docNode)
		if err != nil {
			// Skip invalid YAML documents
			utils.Log.Warningf("Skipping invalid YAML document in %s: %v", manifestFile, err)
			updatedDocuments = append(updatedDocuments, docBytes) // Keep original
			continue
		}

		// Skip empty documents
		if docNode.Kind == 0 || len(docNode.Content) == 0 {
			updatedDocuments = append(updatedDocuments, docBytes)
			continue
		}

		// Process this document
		docUpdated := d.updateNamespaceInDocument(&docNode)
		if docUpdated {
			updated = true
		}

		// Serialize the document back to YAML
		var buf bytes.Buffer
		enc := yaml.NewEncoder(&buf)
		enc.SetIndent(2)
		err = enc.Encode(&docNode)
		enc.Close()
		if err != nil {
			utils.Log.Warningf("Failed to encode document in %s: %v", manifestFile, err)
			updatedDocuments = append(updatedDocuments, docBytes) // Keep original
			continue
		}

		updatedDocuments = append(updatedDocuments, buf.Bytes())
	}

	// If no updates were made, return the original file
	if !updated {
		return manifestFile, nil
	}

	// Create a temporary file for the updated manifest
	tmpFile, err := os.CreateTemp("", "manifest-*.yml")
	if err != nil {
		return "", utils.NewError("failed to create temporary file: %v", err)
	}

	// Join the documents with the --- separator
	separator := []byte("\n---\n")
	for i, doc := range updatedDocuments {
		if i > 0 {
			if _, err := tmpFile.Write(separator); err != nil {
				os.Remove(tmpFile.Name())
				return "", utils.NewError("failed to write document separator: %v", err)
			}
		}
		if _, err := tmpFile.Write(doc); err != nil {
			os.Remove(tmpFile.Name())
			return "", utils.NewError("failed to write document: %v", err)
		}
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpFile.Name())
		return "", utils.NewError("failed to close temporary file: %v", err)
	}

	return tmpFile.Name(), nil
}

// updateNamespaceInDocument updates the namespace in a single YAML document
// Returns true if the document was updated
func (d *CustomDeployer) updateNamespaceInDocument(doc *yaml.Node) bool {
	// Only process mapping nodes (key-value pairs)
	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return false
	}

	// Get the root mapping node
	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return false
	}

	// Look for metadata field in the mapping
	metadataNode := findChildByKey(root, "metadata")
	if metadataNode == nil || metadataNode.Kind != yaml.MappingNode {
		return false
	}

	// Look for namespace field in metadata
	namespaceNode := findChildByKey(metadataNode, "namespace")
	if namespaceNode != nil {
		// Update existing namespace
		if namespaceNode.Value != d.Config.Namespace {
			oldNamespace := namespaceNode.Value
			namespaceNode.Value = d.Config.Namespace
			utils.Log.Infof("Updated namespace from '%s' to '%s'", oldNamespace, d.Config.Namespace)
			return true
		}
		return false
	} else {
		// Add namespace field if not found
		keyNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: "namespace",
		}
		valueNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: d.Config.Namespace,
		}
		metadataNode.Content = append(metadataNode.Content, keyNode, valueNode)
		utils.Log.Infof("Added namespace '%s'", d.Config.Namespace)
		return true
	}
}

// findChildByKey returns the child node with the given key in a mapping node
// Returns nil if the key is not found
func findChildByKey(mappingNode *yaml.Node, key string) *yaml.Node {
	if mappingNode.Kind != yaml.MappingNode {
		return nil
	}

	for i := 0; i < len(mappingNode.Content); i += 2 {
		if i+1 < len(mappingNode.Content) && mappingNode.Content[i].Value == key {
			return mappingNode.Content[i+1]
		}
	}
	return nil
}
