<script lang="ts">
    import Button,{ Label } from '@smui/button'
    import Dialog,{ Actions,Content,Title } from '@smui/dialog'
    import IconButton from '@smui/icon-button/src/IconButton.svelte'
    import Progress from '../../../../components/Progress.svelte'
    import { createSnackbar } from '../../../../global'

    let loading = false
    export let open = false
    export let name = ''
    export let id = ''

    let img = new Image()
    function loadImage() {
        loading = true
        img.onload = () => {
            loading = false
        }
        img.onerror = () => {
            loading = false
            $createSnackbar(`Video feed of camera '${id}' failed to load`)
        }
        img.src = `/api/camera/feed/${id}?${new Date().getTime()}`
    }
    $: if (open) loadImage()
</script>

<Dialog bind:open aria-labelledby="title" aria-describedby="content" fullscreen>
    <Content id="content">
        <Title id="title">{name}</Title>
        <Progress id="loader" bind:loading />
        <div class="img__wrapper">
            <img bind:this={img} alt="video feed of camera" />
        </div>
    </Content>
    <Actions>
        <IconButton
            class="material-icons"
            title="Reload"
            on:click={() => {
                loadImage()
            }}>refresh</IconButton
        >
        <Button>
            <Label>Close</Label>
        </Button>
    </Actions>
</Dialog>

<style lang="scss">
    @use '../../../../_mixins.scss' as *;
    .img__wrapper {
        display: flex;
        justify-content: center;
        align-items: center;
    }
    img {
        height: 100%;
        width: 100%;
        object-fit: cover;

        @include mobile {
            height: min-content;
        }

        @include widescreen {
            height: available;
            width: min-content;
        }
    }
</style>
