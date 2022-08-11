<script lang="ts">
    import IconButton from "@smui/icon-button";

    import Page from "../../Page.svelte";
    import Logs from "./logEvents/Logs.svelte";
    import Progress from "../../components/Progress.svelte";
    import Button from "@smui/button";

    // Specifies whether the log event dialog should be visible or not
    let logsOpen = false;
</script>

<Logs />

<Logs bind:open={logsOpen} />

<Page>
    <div id="header" class="mdc-elevation--z4">
        <h6>System Configuration</h6>
        <div id="header__buttons">
            <IconButton title="Refresh" class="material-icons"
                >refresh</IconButton
            >
            <Button on:click={() => (logsOpen = true)}>Logs</Button>
        </div>
    </div>
    <Progress id="loader" loading={false} />
    <div id="content">
        <div id="left" class="mdc-elevation--z1">Some content here</div>
        <div id="logs" class="mdc-elevation--z1">
            <Logs />
        </div>
    </div></Page
>

<style lang="scss">
    @use "../../mixins" as *;

    #header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 0.1rem 1.3rem;
        box-sizing: border-box;
        background-color: var(--clr-height-1-4);
        min-height: 3.5rem;

        &__buttons {
            display: flex;
            align-items: center;
        }

        h6 {
            margin: 0.5em 0;
            @include mobile {
                // Hide title on mobile due to space limitations
                display: none;
            }
        }
    }

    #content {
        padding: 1rem 1.5rem;
        box-sizing: border-box;
        flex-direction: column;
        display: flex;
        gap: 1rem;

        @include widescreen {
            height: calc(100vh - 60px);
            flex-direction: row;
        }

        @include mobile {
            min-height: calc(100vh - 48px - 3.5rem);
            padding: 1rem;
        }

        #left {
            background-color: var(--clr-height-0-1);
            border-radius: 0.4rem;
            height: 75%;
            width: 100%;

            @include widescreen {
                height: 100%;
                width: 80%;
            }
        }

        #logs {
            background-color: var(--clr-height-0-1);
            border-radius: 0.4rem;
            height: 25%;
            width: 100%;

            @include widescreen {
                height: 100%;
                width: 20%;
            }
        }
    }
</style>
