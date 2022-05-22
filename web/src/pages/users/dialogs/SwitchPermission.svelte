<script lang="ts">
    import FormField from '@smui/form-field/'
    import Switch from '@smui/switch'
    import Progress from '../../../components/Progress.svelte'
    import { sleep } from '../../../global'

    export let name: string
    export let roomId: string
    export let id: string

    export let active // Shows if the user currently has the permission
    let loading = false

    export let grantFunc: (_: string) => {}
    export let removeFunc: (_: string) => {}

    let loadingStart: number = Date.now()
    let loadingTime: number

    // Handle switch updates
    async function toggle(event: CustomEvent<{ selected: boolean }>) {
        loadingStart = Date.now()
        loading = true
        updateLoadingTime()
        try {
            if (event.detail.selected) {
                await grantFunc(id)
            } else await removeFunc(id)
        } catch (err) {
            await sleep(1000)
            active = !active
        }
        loading = false
    }

    // Calculate the time spent waiting for the serve'rs response
    // Is used in the switch to disabled it when the time is greater than 50ms
    // Prevents ugly flickering and stops the user from clicking multiple times causing interference
    async function updateLoadingTime() {
        if (loading) {
            loadingTime = (loadingStart - Date.now()) * -1
            await sleep(5)
            updateLoadingTime()
        }
    }
</script>

<div class="permission mdc-elevation--z3">
    <div class="top">
        <h6>{name}</h6>
        <div>
            <Progress type="circular" bind:loading />
            <pre><span>{id}</span>/{roomId}</pre>
        </div>
    </div>
    <FormField>
        <Switch
            on:SMUISwitch:change={toggle}
            checked={active}
            disabled={loading && loadingTime > 50}
        />
        <span slot="label">Switch {active ? 'granted' : 'denied'}</span>
    </FormField>
</div>

<style lang="scss">
    @use '../../../mixins' as *;
    .permission {
        width: 100%;
        min-height: 6rem;
        display: flex;
        padding: 1rem;
        border-radius: 0.3rem;
        display: flex;
        flex-direction: column;
        justify-content: start;
        background-color: var(--clr-height-1-3);

        span {
            color: var(--clr-text);
        }

        @include widescreen {
            width: 21%;
        }
    }
    .top {
        display: flex;
        justify-content: space-between;
        align-items: center;
        flex-wrap: wrap;

        @include widescreen {
            align-items: flex-start;
        }

        h6 {
            margin: 0;
            margin-bottom: 0.5rem;
            color: var(--clr-text);
            max-width: 10rem;
            word-break: keep-all;
            overflow: hidden;
            text-overflow: ellipsis;

            @include widescreen {
                max-width: 100%;
            }

            @include mobile {
                max-width: 70vw;
            }
        }

        div {
            display: flex;
            align-items: center;

            @include widescreen {
                flex-direction: row-reverse;
            }

            @include mobile {
                gap: 1rem;
                flex-wrap: wrap;
                flex-direction: row-reverse;
            }

            pre {
                font-size: 0.8rem;
                background-color: var(--clr-height-3-6);
                padding: 0.1rem 0.2rem;
                border-radius: 0.1rem;
                max-width: 20rem;
                overflow: hidden;
                text-overflow: ellipsis;

                span {
                    color: var(--clr-text);
                }

                @include widescreen {
                    display: block;
                    max-width: 13rem;
                }

                @include mobile {
                    max-width: 50vw;
                }
            }
        }
    }
</style>
