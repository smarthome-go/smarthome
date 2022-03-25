#!/usr/bin/env python3
# This file is used to update the version number in all relevant places
# The SemVer (https://semver.org) versioning system is used.
import re

docker_motd_path = "docker/motd"
package_json_path = "web/package.json"
main_go_path = "main.go"
readme_path = "README.md"

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
    
# The NPM `package.json`
with open(package_json_path, "r") as package_json:
    content = package_json.read()
    old_version = content.split("\"version\": \"")[1].split("\",\n")[0]
    print(f"Found old version in {package_json_path}:", old_version)

with open(package_json_path, "w") as package_json:
    package_json.write(content.replace(old_version, VERSION))
    
# The NPM `package.json`
with open(readme_path, "r") as readme:
    content = readme.read()
    old_version = content.split("### Version: `")[1].split("`\n")[0]
    print(f"Found old version in {readme_path}:", old_version)
    
with open(readme_path, "w") as readme:
    readme.write(content.replace(old_version, VERSION))
    
print(f"Version has been upgraded from '{old_version}' to '{VERSION}'")
