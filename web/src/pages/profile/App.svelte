<script lang="ts">
    import IconButton from "@smui/icon-button";
    import Progress from "../../components/Progress.svelte";
    import Page from "../../Page.svelte";
import Preview from "./Preview.svelte";
</script>

<Page>
    <div id="header" class="mdc-elevation--z4">
        <h6>Your Profile</h6>
        <div id="header__buttons">
            <IconButton title="Refresh" class="material-icons"
                >refresh</IconButton
            >
        </div>
    </div>
    <Progress id="loader" loading={false} />
    <div id="content">
        <div id="preview" class="mdc-elevation--z1">
            <Preview></Preview>
        </div>
        <div id="inputs" class="mdc-elevation--z1" />
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
        height: calc(100vh - 60px);
        padding: 1rem 1.5rem;
        display: flex;
        gap: 1rem;
        flex-direction: column;
        box-sizing: border-box;

        @include widescreen {
            flex-direction: row;
        }

        @include mobile {
            min-height: calc(100vh - 48px - 3.5rem);
        }
    }

    #preview,
    #inputs {
        background-color: var(--clr-height-0-1);
        height: 10rem;

        @include widescreen {
            width: 50%;
            height: 100%;
        }
    }
</style>
