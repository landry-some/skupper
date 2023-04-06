package main

import (
	"fmt"

	"github.com/skupperproject/skupper/api/types"
	"github.com/skupperproject/skupper/pkg/domain"
	"github.com/skupperproject/skupper/pkg/domain/podman"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes/scheme"
)

type SkupperPodmanLink struct {
	podman      *SkupperPodman
	linkHandler domain.LinkHandler
}

func (s *SkupperPodmanLink) Create(cmd *cobra.Command, args []string) error {
	// reading secret from file
	var secret corev1.Secret
	serializer := json.NewSerializerWithOptions(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme, json.SerializerOptions{Yaml: true})
	_, _, err := serializer.Decode(connectorCreateOpts.Yaml, nil, &secret)
	if err != nil {
		return fmt.Errorf("error decoding token - %w", err)
	}

	linkHandler := podman.NewLinkHandlerPodman(s.podman.currentSite, s.podman.cli)
	return linkHandler.Create(&secret, connectorCreateOpts.Name, int(connectorCreateOpts.Cost))
}

func (s *SkupperPodmanLink) CreateFlags(cmd *cobra.Command) {}

func (s *SkupperPodmanLink) Delete(cmd *cobra.Command, args []string) error {
	linkHandler := podman.NewLinkHandlerPodman(s.podman.currentSite, s.podman.cli)
	return linkHandler.Delete(connectorRemoveOpts.Name)
}

func (s *SkupperPodmanLink) DeleteFlags(cmd *cobra.Command) {}

func (s *SkupperPodmanLink) List(cmd *cobra.Command, args []string) error {
	return notImplementedErr
}

func (s *SkupperPodmanLink) ListFlags(cmd *cobra.Command) {}

func (s *SkupperPodmanLink) Status(cmd *cobra.Command, args []string) error {
	return notImplementedErr
}

func (s *SkupperPodmanLink) StatusFlags(cmd *cobra.Command) {}

func (s *SkupperPodmanLink) NewClient(cmd *cobra.Command, args []string) {
	s.podman.NewClient(cmd, args)
}

func (s *SkupperPodmanLink) Platform() types.Platform {
	return s.podman.Platform()
}

func (s *SkupperPodmanLink) LinkHandler() domain.LinkHandler {
	if s.linkHandler != nil {
		return s.linkHandler
	}
	if s.podman.currentSite == nil {
		return nil
	}
	sitePodman := s.podman.currentSite
	s.linkHandler = podman.NewLinkHandlerPodman(sitePodman, s.podman.cli)
	return s.linkHandler
}
