<script lang="ts">
    import Snackbar, { Label, Actions } from '@smui/snackbar'
    import IconButton from '@smui/icon-button'
    import { infoBar } from '../../global'
    import NavBar from '../../components/NavBar.svelte'
</script>

<NavBar />
<main>
    Dashboard
</main>
<Snackbar bind:this={$infoBar.bar}>
    <Label>{$infoBar.message}</Label>
    <Actions>
        <IconButton class="material-icons" title="Dismiss">close</IconButton>
    </Actions>
</Snackbar>

<style lang="scss">
    @use '../../mixins' as *;

    main {
        margin-left: 5.125rem;
        transition: margin-left .3s;
        @include mobile {
            margin-left: 0;
            margin-top: 3.5rem;
        }
        @include widescreen {
            margin-left: 13rem;
        }
    }
</style>
