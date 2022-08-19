<script lang="ts">
    import Progress from "../../components/Progress.svelte";

    // Whether the internal progress indicator should be active
    export let loading = false;
</script>

<div class="box mdc-elevation--z2">
    <div class="box__header mdc-elevation--z3">
        <div class="box__header__left">
            <slot name="header" />
        </div>
        <div class="box__header__right">
            <slot name="header-right" />
            <Progress bind:loading type="circular" />
        </div>
    </div>
    <div class="box__content">
        <slot name="content" />
    </div>
</div>

<style lang="scss">
    @use "../../_mixins.scss" as *;

    .box {
        width: 30rem;
        height: 22rem;
        flex-grow: 1;

        background-color: var(--clr-height-0-2);
        border-radius: 0.3rem;
        overflow-x: hidden;

        &__header {
            display: flex;
            align-items: center;
            justify-content: space-between;
            background-color: var(--clr-height-2-3);
            height: 3rem;
            padding: 0 1rem;

            &__left {
                // Typography
                color: var(--clr-text-hint);
                font-weight: bold;
            }

            &__right {
                display: flex;
                align-items: center;
                justify-content: space-between;
            }
        }
        &__content {
            padding: 0.75rem 1.25rem;
            height: calc(100% - 3rem);
            box-sizing: border-box;
            overflow-y: auto;

            @include mobile {
                padding: 0.9rem 0.9rem;
                overflow: visible;
            }
        }
    }
</style>
