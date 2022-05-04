<script lang="ts">
    import Button, { Label } from "@smui/button";
    import { onMount } from "svelte";
    import Dialog, {
        Actions,
        Content,
        InitialFocus,
        Title,
    } from "@smui/dialog";
    import Textfield from "@smui/textfield";
    import CharacterCounter from "@smui/textfield/character-counter";
    import { createSnackbar } from "../../../../global";
    import IconButton from "@smui/icon-button/src/IconButton.svelte";
    import Progress from "../../../../components/Progress.svelte";

    let loading = false;

    export let open = false;
    export let name = "";
    export let id = "";

    let img = new Image();
    function loadImage() {
        loading = true;
        img.onload = () => {
            loading = false;
        };
        img.onerror = (err) => {
            loading = false;
            $createSnackbar(`Video feed of camera '${id}' failed to load`);
        };
        img.src = `/api/camera/feed/${id}?${new Date().getTime()}`;
    }
    $: if (open) loadImage();
</script>

<Dialog bind:open aria-labelledby="title" aria-describedby="content" fullscreen>
    <Content id="content">
        <Title id="title">{name}</Title>
        <Progress id="loader" bind:loading />
        <img bind:this={img} alt="video feed of camera" />
    </Content>
    <Actions>
        <IconButton
            class="material-icons"
            title="Reload"
            on:click={() => {
                loadImage();
            }}>refresh</IconButton
        >
        <Button>
            <Label>Close</Label>
        </Button>
    </Actions>
</Dialog>

<style lang="scss">
    @use '../../../../_mixins.scss' as *;
    img {
        width: 100%;
        height: 100%;
        object-fit: cover;

        @include mobile {
            height: min-content;
        }
    }
</style>
