<script lang='ts'>
    import IconButton from "@smui/icon-button";
    import Progress from '../../components/Progress.svelte'
    import Ripple from '@smui/ripple'
    import { createEventDispatcher } from "svelte";

    export let hasEditPermission = false
    let isWide = hasEditPermission
    export let isTall = false

    export let loading = false

    export let name = ""

    const dispatch = createEventDispatcher()

    function showInfo() {
        dispatch('info_show')
    }

    function showEdit() {
        dispatch('edit_show')
    }
</script>

<div class="device mdc-elevation--z3" class:wide={isWide} class:tall={isTall}>
    <div class="device__top">
        <slot name="top"></slot>
        <div
            class="device__top__left"
            use:Ripple={{ surface: true }}
            on:click={showInfo}
            on:keydown={showInfo}
        >
            <span class="device__top__left__name"> {name}</span>
        </div>

        <div class="device__top__right">
            <div>
                <Progress type="circular" bind:loading />
            </div>
            {#if hasEditPermission}
                <IconButton class="material-icons" title="Edit Device" on:click={showEdit}>edit</IconButton
                >
            {/if}
        </div>
    </div>

    <slot name="extend"></slot>
</div>


<style lang="scss">
    @use '../../mixins' as *;
    .device {
        background-color: var(--clr-height-1-3);
        border-radius: 0.3rem;
        width: 15rem;
        height: 3.3rem;
        padding: 0.5rem;
        display: flex;
        flex-direction: column;

        &.tall {
            height: auto;
            min-height: 3.3rem;
        }

        &.wide {
            width: 17rem;

            @include mobile {
                width: 90%;
            }
        }

        &__top {
            display: flex;
            align-items: center;
            justify-content: space-between;

            & > * {
                display: flex;
                align-items: center;
            }

            &__left {
                padding: 2px 5px;
                border-radius: 5px;
                cursor: pointer;
                max-width: 70%;
                gap: 0.2rem;

                &__name {
                    overflow: hidden;
                    text-overflow: ellipsis;
                }
            }

            &__right {

                div {
                    margin-right: 14px;
                    display: flex;
                    align-items: center;
                }
            }
        }


        @include mobile {
            width: 90%;
            height: auto;
            flex-wrap: wrap;
        }
    }
</style>
