"""Kubernetes examples must run AgentHealth safely and stay consistent with the image."""
from pathlib import Path

import yaml

KUBERNETES = Path(__file__).parents[1] / "integrations/kubernetes"
IMAGE = "ghcr.io/theagenthealth/agenthealth"
BINARY = "/usr/local/bin/agenthealth"


def documents():
    for path in sorted(KUBERNETES.glob("*.yaml")):
        if path.name in ("agenthealth.yaml", "kustomization.yaml"):
            continue
        for document in yaml.safe_load_all(path.read_text()):
            yield path.name, document


def pod_specs():
    for name, document in documents():
        spec = document.get("spec", {})
        if document["kind"] == "CronJob":
            spec = spec["jobTemplate"]["spec"]
        if "template" in spec:
            yield name, spec["template"]["spec"]


def test_kustomization_lists_existing_resources_and_pins_image():
    kustomization = yaml.safe_load((KUBERNETES / "kustomization.yaml").read_text())
    for resource in kustomization["resources"]:
        assert (KUBERNETES / resource).is_file(), resource
    images = {image["name"]: image for image in kustomization["images"]}
    assert images[IMAGE]["newTag"] not in ("", "latest")


def test_agenthealth_containers_are_restricted():
    found = 0
    for name, spec in pod_specs():
        for container in spec.get("initContainers", []) + spec["containers"]:
            if container["image"] != IMAGE:
                continue
            found += 1
            assert container["args"][0] in ("check", "doctor"), name
            context = container["securityContext"]
            assert context["runAsNonRoot"] and context["readOnlyRootFilesystem"], name
            assert context["allowPrivilegeEscalation"] is False, name
            assert context["capabilities"]["drop"] == ["ALL"], name
            assert context["seccompProfile"]["type"] == "RuntimeDefault", name
            assert all(mount.get("readOnly") for mount in container["volumeMounts"]), name
    assert found == 3


def test_dependency_checks_never_drive_liveness():
    probes = 0
    for name, spec in pod_specs():
        for container in spec["containers"]:
            assert BINARY not in str(container.get("livenessProbe", "")), name
            for probe in ("startupProbe", "readinessProbe"):
                if probe in container:
                    probes += 1
                    assert container[probe]["exec"]["command"][:2] == [BINARY, "check"], name
                    assert container[probe]["timeoutSeconds"] >= 5, name
    assert probes == 2


def test_jobs_do_not_retry_or_overlap():
    for name, document in documents():
        if document["kind"] == "Job":
            assert document["spec"]["backoffLimit"] == 0, name
        if document["kind"] == "CronJob":
            assert document["spec"]["concurrencyPolicy"] == "Forbid", name
            assert document["spec"]["jobTemplate"]["spec"]["backoffLimit"] == 0, name
