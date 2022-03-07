# Required
container "region-eu" {
    # Path to the container.
    path = "region-eu.bundle"

    # Idetities for sealing purpose if the container is not sealed.
    identities = [
        "v1.ipk.7u8B1VFrHyMeWyt8Jzj1Nj2BgVB7z-umD8R-OOnJahE", # Security public key
    ]
}

# Optional archive layer
# All files matching the filter will be compressed as a tar.gz and embedded
# in the crate.
archive "production" {
    # Root path from where files are crawled to create the archive.
    root = "./production"

    # Include filters.
    includes = [
        "**"
    ]

    # Exclude filters.
    excludes = [
        "**.go"
    ]
}
