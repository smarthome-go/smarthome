<script lang="ts">
    import IconButton from '@smui/icon-button'
    import { onMount } from 'svelte/internal'
    import Progress from '../../components/Progress.svelte'
    import { createSnackbar,hasPermission,sleep } from '../../global'
    import EditCamera from './dialogs/camera/EditCamera.svelte'
    import ViewCamera from './dialogs/camera/ViewCamera.svelte'
    import type { Camera,SwitchResponse } from './main'

    // Cameras are binded in order to use the editor
    export let cameras: Camera[]

    // Switches are also exported in order to detect if power states change
    export let switches: SwitchResponse[]

    // Reload the image if a switch was changed
    $: if (switches) {
        updateImage()
    }

    // Reloads the image 3 times with a delay of 4s between each iteration
    async function updateImage() {
        if (!loaded) return // If the image is not initially loaded, (first page load), then stop here
        for (let i = 0; i < 3; i++) {
            await sleep(4000)
            await loadImage()
        }
    }

    // Camera metadata
    export let id: string
    export let name: string
    export let url: string

    // Keeps track of dialog state
    let viewOpen = false
    let editOpen = false

    let loading = true
    // Indicates that the fetching of the camera feed is complete
    let loaded = false
    // Indicates wheter the fetching of the camera feed has failed
    let error = false

    // Determines if edit button should be shown
    let hasEditPermission: boolean
    onMount(async () => {
        hasEditPermission = await hasPermission('modifyRooms')
    })

    // Creates an empty image
    let img = new Image()

    // Appends the suffix of the currenty unix-millis to the image's url in order to force a refresh
    // If the image fails to load, a snackbar is created and the `error` boolean is set to `true` 
    async function loadImage() {
        loading = true
        img.onload = () => {
            loaded = true
            loading = false
            error = false
        }
        img.onerror = () => {
            loading = false
            $createSnackbar(`Video feed of camera '${id}' failed to load`)
            error = true
        }
        img.src = `/api/camera/feed/${id}?${new Date().getTime()}`
        while (loading) await sleep(5)
    }

    // Sends a modification request to the server
    async function modifyCamera() {
        loading = true
        try {
            const res = await (
                await fetch('/api/camera/modify', {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ id, name, url }),
                })
            ).json()
            if (!res.success) throw Error(res.error)
            loadImage()
        } catch (err) {
            $createSnackbar(`Could not edit camera: ${err}`)
        }
        loading = false
    }

    // Load the image initially
    onMount(loadImage)
</script>

<!-- Fullscreen camera feed -->
<ViewCamera {id} {name} bind:open={viewOpen} />

<!-- If the user is allowed to modify rooms, mount the edit-camera popup -->
{#if hasEditPermission}
    <EditCamera
        {modifyCamera}
        bind:open={editOpen}
        bind:cameras
        {id}
        bind:name
        bind:url
    />
{/if}

<!-- Actual camera DIV -->
<div class="camera mdc-elevation--z3">
    <!-- Camera feed image, has the class `error` if the stream fails to load -->
    <img
        bind:this={img}
        alt="video feed of camera"
        style:display={error ? 'none' : 'block'}
    />
    <!-- Is loading when the stream fetches -->
    <div class="loader">
        <Progress bind:loading />
    </div>
    <!-- Buttons and texts with a transparent background, serves as the overlay -->
    <div class="over" class:blur={!loaded} class:error>
        {#if loaded || error}
            <div class="over__top">
                <h6>{name}</h6>
                <code>{id}</code>
            </div>
            <!-- Edit-camera button is shown when the user has the permission -->
            <div class="over__buttons">
                {#if hasEditPermission}
                    <IconButton
                        class="material-icons"
                        title="Edit Camera"
                        on:click={() => {
                            editOpen = true
                        }}>edit</IconButton
                    >
                {/if}
                <IconButton
                    class="material-icons"
                    title="Reload"
                    on:click={loadImage}>refresh</IconButton
                >
                <IconButton
                    class="material-icons"
                    title="View Camera"
                    on:click={() => {
                        viewOpen = true
                    }}>preview</IconButton
                >
            </div>
        {/if}
    </div>
</div>

<style lang="scss">
    @use '../../mixins' as *;
    .camera {
        height: 9rem;
        min-width: 16rem;
        background-color: var(--clr-height-1-3);
        position: relative;
        border-radius: 0.4rem;
        overflow: hidden;
        flex-shrink: 0;
    }
    .loader {
        width: 100%;
        height: min-content;
        z-index: 10;
        position: absolute;
    }
    img {
        height: 100%;
        width: 100%;
        position: absolute;
        object-fit: cover;
    }
    .over {
        backdrop-filter: blur(20px);
        border-radius: 0.4rem;
        background: linear-gradient(
            45deg,
            rgba(0, 0, 0, 0.2),
            rgba(0, 0, 0, 0.4)
        );
        width: 100%;
        height: 100%;
        position: absolute;
        opacity: 0;
        transition: opacity 0.2s ease-out;
        padding: 1rem;
        box-sizing: border-box;
        display: flex;
        flex-direction: column;
        justify-content: space-between;

        &.error {
            border: 0.1rem solid var(--clr-error);
            background: linear-gradient(
                45deg,
                rgba(100, 100, 100, 0.1),
                rgba(0, 0, 0, 0.3)
            );
        }
        &:hover {
            opacity: 1;
        }
        h6 {
            margin: 0;
        }
        @include mobile {
            opacity: 1;
            backdrop-filter: none;
            background: none;
            padding: 0.5rem;
            &__top {
                display: flex;
                h6 {
                    padding: 0 1rem;
                    font-size: 1rem;
                    backdrop-filter: blur(10px);
                    border-radius: 0.2rem;
                    background: linear-gradient(
                        45deg,
                        rgba(0, 0, 0, 0.1),
                        rgba(0, 0, 0, 0.3)
                    );
                }
                code {
                    display: none;
                }
            }
            &__buttons {
                backdrop-filter: blur(10px);
                border-radius: 0.5rem;
                background: linear-gradient(
                    45deg,
                    rgba(0, 0, 0, 0.1),
                    rgba(0, 0, 0, 0.3)
                );
            }
        }
    }
    .blur {
        opacity: 1;
    }
</style>
