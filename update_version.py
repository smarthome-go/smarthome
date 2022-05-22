#!/usr/bin/env python3
# This file is used to update the version number in all relevant places
# The SemVer (https://semver.org) versioning system is used.
import re

docker_motd_path = "docker/container/motd"
dockerfile_path = "docker/container/Dockerfile"
package_json_path = "web/package.json"
main_go_path = "main.go"
readme_path = "README.md"
changelog_path = "./docs/CHANGELOG.md"
makefile_path = "Makefile"
docker_compose_path = "./docker-compose.yml"

with open(main_go_path, "r") as main_go:
    content = main_go.read()
    old_version = content.split("utils.Version = \"")[1].split("\"\n")[0]
    print(f"Found old version in {main_go_path}:", old_version)

try:
    VERSION = input(
        f"Current version: {old_version}\nNew version (without 'v' prefix): ")
except KeyboardInterrupt:
    print("\nCanceled by user")
    quit()

if VERSION == "":
    VERSION = old_version

if not re.match(r"^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$", VERSION):
    print(
        f"\x1b[31mThe version: '{VERSION}' is not a valid SemVer version.\x1b[0m")
    quit()


with open(main_go_path, "w") as main_go:
    main_go.write(content.replace(old_version, VERSION))

# The Docker motd
with open(docker_motd_path, "r") as docker_motd:
    content = docker_motd.read()
    old_version = content.split("Version  : ")[1].split("\n")[0]
    print(f"Found old version in {docker_motd_path}:", old_version)

with open(docker_motd_path, "w") as main_go:
    main_go.write(content.replace(old_version, VERSION))

# The Dockerfile
with open(dockerfile_path, "r") as dockerfile:
    content = dockerfile.read()
    old_version = content.split("LABEL version=\"")[1].split("\"\n")[0]
    print(f"Found old version in {dockerfile_path}:", old_version)

with open(dockerfile_path, "w") as dockerfile:
    dockerfile.write(content.replace(old_version, VERSION))

# The NPM `package.json`
with open(package_json_path, "r") as package_json:
    content = package_json.read()
    old_version = content.split("\"version\": \"")[1].split("\",\n")[0]
    print(f"Found old version in {package_json_path}: {old_version}")

with open(package_json_path, "w") as package_json:
    package_json.write(content.replace(f'"version": "{old_version}', f'"version": "{VERSION}'))

# The `README.md`
with open(readme_path, "r") as readme:
    content = readme.read()
    old_version = content.split("**Version**: `")[1].split("`\n")[0]
    print(f"Found old version in {readme_path}: {old_version}")

with open(readme_path, "w") as readme:
    readme.write(content.replace(old_version, VERSION))

# The `CHANGELOG.md`
with open(changelog_path, "r") as changelog:
    content = changelog.read()
    old_version = content.split("## Changelog for v")[1].split("\n")[0]
    print(f"Found old version in {changelog_path}: {old_version}")

with open(changelog_path, "w") as changelog:
    changelog.write(content.replace(old_version, VERSION))

# The `Makefile`
with open(makefile_path, "r") as makefile:
    content = makefile.read()
    old_version = content.split("version := ")[1].split("\n")[0]
    print(f"Found old version in {makefile_path}: {old_version}")

with open(makefile_path, "w") as makefile:
    makefile.write(content.replace(old_version, VERSION))

# The `docker_compose.yml`
with open(docker_compose_path, "r") as compose:
    content = compose.read()
    old_version = content.split(
        "image: mikmuellerdev/smarthome:")[1].split("\n")[0]
    print(f"Found old version in {docker_compose_path}:", old_version)

with open(docker_compose_path, "w") as compose:
    compose.write(content.replace(old_version, VERSION))

print(f"Version has been changed from '{old_version}' -> '{VERSION}'")
