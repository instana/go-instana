// (c) Copyright IBM Corp. 2022

package recipes

import (
	"github.com/instana/go-instana/registry"
	"go/ast"
	"go/token"
)

func init() {
	registry.Default.Register("github.com/Shopify/sarama", NewSarama())
}

// NewSarama returns Sarama recipe
func NewSarama() *Sarama {
	return &Sarama{InstanaPkg: "instasarama", defaultRecipe: defaultRecipe{}}
}

// Sarama instruments github.com/Shopify/sarama package with Instana
type Sarama struct {
	InstanaPkg    string
	defaultRecipe defaultRecipe
}

// ImportPath returns instrumentation import path
func (recipe *Sarama) ImportPath() string {
	return "github.com/instana/go-sensor/instrumentation/instasarama"
}

// Instrument applies recipe to the ast Node
func (recipe *Sarama) Instrument(fset *token.FileSet, f ast.Node, targetPkg, sensorVar string) (changed bool) {
	return recipe.defaultRecipe.instrument(fset, f, targetPkg, sensorVar, recipe.InstanaPkg, recipe.ImportPath(), map[string]insertOption{
		"NewAsyncProducer":           {sensorPosition: lastInsertPosition},
		"NewAsyncProducerFromClient": {sensorPosition: lastInsertPosition},
		"NewConsumer":                {sensorPosition: lastInsertPosition},
		"NewConsumerFromClient":      {sensorPosition: lastInsertPosition},
		"NewSyncProducer":            {sensorPosition: lastInsertPosition},
		"NewSyncProducerFromClient":  {sensorPosition: lastInsertPosition},
		"NewConsumerGroup":           {sensorPosition: lastInsertPosition},
		"NewConsumerGroupFromClient": {sensorPosition: lastInsertPosition},
	})
}
