<script lang="ts">
    import Button, { Label } from "@smui/button";
    import Dialog, {
        Actions,
        Content,
        InitialFocus,
        Title,
    } from "@smui/dialog";
    import { createEventDispatcher } from "svelte";
    import { createSnackbar } from "../../../global";

    // Create event dispatcher
    const dispatch = createEventDispatcher();

    // Is bound externally and provides an abstraction over the dialog's behaviour
    export let open = false;
    $: if (open) {
        // Automatically open the file picker when the dialog is opened from the outsite
        fileInput.click();
        // Set the outer API to false again
        open = false;
    }

    // Keeps track of whether the dialog should be open or closed
    let dialogOpen = false;
    $: if (!dialogOpen) open = false;

    // Avatar / image variables
    let avatarSrc = "";
    let avatarImage: File = undefined;
    let fileInput: HTMLInputElement = undefined;

    // Preview DIV element
    let preview: HTMLDivElement = undefined;

    // Callback to be executed as soon as a file has been picked
    function onFileSelected(e: Event) {
        avatarImage = (e.target as HTMLInputElement).files[0];
        if (avatarImage === undefined) {
            open = false;
            dialogOpen = false;
            return;
        }
        const reader = new FileReader();
        reader.readAsDataURL(avatarImage);
        reader.onload = (e) => {
            avatarSrc = e.target.result as string;

            // Set the preview image to display the selected image
            preview.style.backgroundImage = `url(${avatarSrc})`;
            // Trigger the internal dialog opening
            dialogOpen = true;
        };
    }

    // Sends the newly selected file to the server
    async function submitImage() {
        try {
            let formData = new FormData();
            formData.append("file", avatarImage);

            const res = await (
                await fetch("/api/user/avatar/upload", {
                    method: "POST",
                    body: formData,
                })
            ).json();

            // Check for possible errors
            if (!res.success) throw Error(res.error);
            $createSnackbar("Successfully updated avatar");
            // Dispatch a `update` event to the parent component
            // The avatar's source is passed as a URL-source string
            dispatch("update", avatarSrc);
        } catch (err) {
            $createSnackbar(`Failed to upload avatar: ${err}`);
        }
        // Close the dialog again
        open = false;
        dialogOpen = false;
    }
</script>

<Dialog
    bind:open={dialogOpen}
    aria-labelledby="simple-title"
    aria-describedby="simple-content"
>
    <Title id="simple-title">Confirm Avatar Upload</Title>
    <Content id="simple-content">
        <div id="content">
            <div id="preview" bind:this={preview} />
            <span class="text-hint">
                <strong>Note: </strong>
                It might take up to
                <span style="color: var(--clr-primary);">6 hours</span> for your
                image to appear on every device. If you are impatient, try clearing
                your browser's cache or force-reloading this page.
            </span>
        </div>
        <input
            style="display:none"
            type="file"
            accept=".jpg, .jpeg, .png"
            on:input={(e) => onFileSelected(e)}
            bind:this={fileInput}
        />
    </Content>
    <Actions>
        <Button>
            <Label>Cancel</Label>
        </Button>
        <Button on:click={submitImage} defaultAction use={[InitialFocus]}>
            <Label>Submit</Label>
        </Button>
    </Actions>
</Dialog>

<style lang="scss">
    @use "../../../mixins" as *;
    #content {
        display: flex;
        gap: 1rem;
        align-items: center;
        margin-top: 1rem;

        @include mobile {
            flex-direction: column;
            margin-top: 0;
        }
    }
    #preview {
        background-position: center;
        background-size: cover;
        background-repeat: no-repeat;
        border-radius: 50%;
        aspect-ratio: 1;
        height: 8rem;

        @include mobile {
            margin-top: 1rem;
            margin-bottom: 1rem;
        }
    }
</style>
