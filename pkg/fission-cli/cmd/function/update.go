/*
Copyright 2019 The Fission Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package function

import (
	"fmt"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	fv1 "github.com/fission/fission/pkg/apis/core/v1"
	"github.com/fission/fission/pkg/fission-cli/cliwrapper/cli"
	"github.com/fission/fission/pkg/fission-cli/cmd"
	_package "github.com/fission/fission/pkg/fission-cli/cmd/package"
	"github.com/fission/fission/pkg/fission-cli/cmd/spec"
	"github.com/fission/fission/pkg/fission-cli/console"
	flagkey "github.com/fission/fission/pkg/fission-cli/flag/key"
	"github.com/fission/fission/pkg/fission-cli/util"
)

type UpdateSubCommand struct {
	cmd.CommandActioner
	function *fv1.Function
	specFile string
}

func Update(input cli.Input) error {
	return (&UpdateSubCommand{}).do(input)
}

func (opts *UpdateSubCommand) do(input cli.Input) error {
	err := opts.complete(input)
	if err != nil {
		return err
	}
	return opts.run(input)
}

func (opts *UpdateSubCommand) complete(input cli.Input) error {
	fnName := input.String(flagkey.FnName)
	_, fnNamespace, err := opts.GetResourceNamespace(input, flagkey.NamespaceFunction)
	if err != nil {
		return fmt.Errorf("error in updating function : %w", err)
	}
	if input.Bool(flagkey.SpecSave) {
		opts.specFile = fmt.Sprintf("function-%s.yaml", fnName)
	}

	function, err := opts.Client().FissionClientSet.CoreV1().Functions(fnNamespace).Get(input.Context(), input.String(flagkey.FnName), metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("read function '%v': %w", fnName, err)
	}

	envName := input.String(flagkey.FnEnvironmentName)
	// if the new env specified is the same as the old one, no need to update package
	// same is true for all update parameters, but, for now, we don't check all of them - because, its ok to
	// re-write the object with same old values, we just end up getting a new resource version for the object.
	if len(envName) > 0 && envName == function.Spec.Environment.Name {
		envName = ""
	}

	pkgName := input.String(flagkey.FnPackageName)
	entrypoint := input.String(flagkey.FnEntrypoint)

	secretNames := input.StringSlice(flagkey.FnSecret)
	cfgMapNames := input.StringSlice(flagkey.FnCfgMap)

	var secrets []fv1.SecretReference
	var configMaps []fv1.ConfigMapReference

	if len(secretNames) > 0 {

		// check that the referenced secret is in the same ns as the function, if not give a warning.
		for _, secretName := range secretNames {
			err := util.SecretExists(input.Context(), &metav1.ObjectMeta{Namespace: fnNamespace, Name: secretName}, opts.Client().KubernetesClient)
			if k8serrors.IsNotFound(err) {
				console.Warn(fmt.Sprintf("secret %s not found in Namespace: %s. Secret needs to be present in the same namespace as function", secretName, fnNamespace))
			}
		}

		for _, secretName := range secretNames {
			newSecret := fv1.SecretReference{
				Name:      secretName,
				Namespace: fnNamespace,
			}
			secrets = append(secrets, newSecret)
		}

		function.Spec.Secrets = secrets
	}

	if len(cfgMapNames) > 0 {

		// check that the referenced cfgmap is in the same ns as the function, if not give a warning.
		for _, cfgMapName := range cfgMapNames {
			err := util.ConfigMapExists(input.Context(), &metav1.ObjectMeta{Namespace: fnNamespace, Name: cfgMapName}, opts.Client().KubernetesClient)
			if k8serrors.IsNotFound(err) {
				console.Warn(fmt.Sprintf("ConfigMap %s not found in Namespace: %s. ConfigMap needs to be present in the same namespace as the function", cfgMapName, fnNamespace))
			}
		}

		for _, cfgMapName := range cfgMapNames {
			newCfgMap := fv1.ConfigMapReference{
				Name:      cfgMapName,
				Namespace: fnNamespace,
			}
			configMaps = append(configMaps, newCfgMap)
		}
		function.Spec.ConfigMaps = configMaps
	}

	if len(envName) > 0 {
		function.Spec.Environment.Name = envName
	}

	if len(entrypoint) > 0 {
		function.Spec.Package.FunctionName = entrypoint
	}

	if input.IsSet(flagkey.FnExecutionTimeout) {
		fnTimeout := input.Int(flagkey.FnExecutionTimeout)
		if fnTimeout <= 0 {
			return fmt.Errorf("--%v must be greater than 0", flagkey.FnExecutionTimeout)
		}
		function.Spec.FunctionTimeout = fnTimeout
	}

	if input.IsSet(flagkey.FnIdleTimeout) {
		fnTimeout := input.Int(flagkey.FnIdleTimeout)
		function.Spec.IdleTimeout = &fnTimeout
	}

	err = checkExecutorPoolManager(input, function.Spec.InvokeStrategy.ExecutionStrategy.ExecutorType)
	if err != nil {
		return err
	}

	if input.IsSet(flagkey.FnConcurrency) {
		function.Spec.Concurrency = input.Int(flagkey.FnConcurrency)
	}

	if input.IsSet(flagkey.FnRequestsPerPod) {
		function.Spec.RequestsPerPod = input.Int(flagkey.FnRequestsPerPod)
	}

	if input.IsSet(flagkey.FnRetainPods) {
		function.Spec.RetainPods = input.Int(flagkey.FnRetainPods)
	}

	if input.IsSet(flagkey.FnOnceOnly) {
		function.Spec.OnceOnly = input.Bool(flagkey.FnOnceOnly)
	}
	if len(pkgName) == 0 {
		pkgName = function.Spec.Package.PackageRef.Name
	}

	strategy, err := getInvokeStrategy(input, &function.Spec.InvokeStrategy)
	if err != nil {
		return err
	}
	function.Spec.InvokeStrategy = *strategy

	resReqs, err := util.GetResourceReqs(input, &function.Spec.Resources)
	if err != nil {
		return err
	}

	function.Spec.Resources = *resReqs

	pkg, err := opts.Client().FissionClientSet.CoreV1().Packages(fnNamespace).Get(input.Context(), pkgName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("read package '%v.%v'. Pkg should be present in the same ns as the function: %w", pkgName, fnNamespace, err)
	}

	forceUpdate := input.Bool(flagkey.PkgForce)

	fnList, err := _package.GetFunctionsByPackage(input.Context(), opts.Client(), pkg.ObjectMeta.Name, pkg.ObjectMeta.Namespace)
	if err != nil {
		return fmt.Errorf("error getting function list: %w", err)
	}

	if !forceUpdate && len(fnList) > 1 {
		return fmt.Errorf("package is used by multiple functions, use --%v to force update", flagkey.PkgForce)
	}

	newPkgMeta, err := _package.UpdatePackage(input, opts.Client(), opts.specFile, pkg)
	if err != nil {
		return fmt.Errorf("error updating package '%v': %w", pkgName, err)
	}

	// the package resource version of function has been changed,
	// we need to update function resource version to prevent conflict.
	// TODO: remove this block when deprecating pkg flags of function command.
	if pkg.ObjectMeta.ResourceVersion != newPkgMeta.ResourceVersion {
		var fns []fv1.Function
		// don't update the package resource version of the function we are currently
		// updating to prevent update conflict.
		for _, fn := range fnList {
			if fn.ObjectMeta.UID != function.ObjectMeta.UID {
				fns = append(fns, fn)
			}
		}
		err = _package.UpdateFunctionPackageResourceVersion(input.Context(), opts.Client(), newPkgMeta, fns...)
		if err != nil {
			return fmt.Errorf("error updating function package reference resource version: %w", err)
		}
	}

	// TODO : One corner case where user just updates the pkg reference with fnUpdate, but internally this new pkg reference
	// references a diff env than the spec

	// update function spec with new package metadata
	function.Spec.Package.PackageRef = fv1.PackageRef{
		Namespace:       newPkgMeta.Namespace,
		Name:            newPkgMeta.Name,
		ResourceVersion: newPkgMeta.ResourceVersion,
	}

	if function.Spec.Environment.Name != pkg.Spec.Environment.Name {
		console.Warn("Function's environment is different than package's environment, package's environment will be used for updating function")
		function.Spec.Environment.Name = pkg.Spec.Environment.Name
		function.Spec.Environment.Namespace = pkg.Spec.Environment.Namespace
	}

	opts.function = function

	err = util.ApplyLabelsAndAnnotations(input, &opts.function.ObjectMeta)
	if err != nil {
		return err
	}

	return nil
}

func (opts *UpdateSubCommand) run(input cli.Input) error {
	if input.Bool(flagkey.SpecSave) {
		err := opts.function.Validate()
		if err != nil {
			return fv1.AggregateValidationErrors("Function", err)
		}
		err = spec.SpecSave(*opts.function, opts.specFile, false)
		if err != nil {
			return fmt.Errorf("error saving function spec: %w", err)
		}
		return nil
	}
	_, err := opts.Client().FissionClientSet.CoreV1().Functions(opts.function.Namespace).Update(input.Context(), opts.function, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("error updating function: %w", err)
	}

	fmt.Printf("Function '%v' updated\n", opts.function.ObjectMeta.Name)
	return nil
}
