<script lang="ts">
    import Button,{ Label } from '@smui/button'
    import Dialog,{
    Actions,
    Content,
    Header,
    InitialFocus,
    Title
    } from '@smui/dialog'
    import IconButton from '@smui/icon-button'
    import {
    automations,
    hmsLoaded,
    homescripts,
    parseCronExpressionToTime
    } from '../main'

    interface day {
        short: string
        long: string
        index: number
    }

    const days: day[] = [
        {
            index: 0,
            long: 'Sunday',
            short: 'su',
        },
        {
            index: 1,
            long: 'Monday',
            short: 'mo',
        },
        {
            index: 2,
            long: 'Tuesday',
            short: 'tu',
        },
        {
            index: 3,
            long: 'Wednesday',
            short: 'we',
        },
        {
            index: 4,
            long: 'Thursday',
            short: 'th',
        },
        {
            index: 5,
            long: 'Friday',
            short: 'fr',
        },
        {
            index: 6,
            long: 'Saturday',
            short: 'sa',
        },
    ]

    export let open = false
</script>

<Dialog
    bind:open
    aria-labelledby="title"
    aria-describedby="content"
    fullscreen={$hmsLoaded && $homescripts.length > 0}
>
    <Header>
        <Title id="title">Automation Overview</Title>
        {#if $hmsLoaded && $homescripts.length > 0}
            <IconButton action="close" class="material-icons">close</IconButton>
        {/if}
    </Header>
    <Content id="content">
        <div id="days">
            {#each days as day (day.short)}
                <div class="day">
                    <div class="day__header">
                        {day.long}
                    </div>
                    <div class="day__automations">
                        {#each $automations
                            .filter( (a) => parseCronExpressionToTime(a.cronExpression).days.includes(day.index) )
                            .sort((a, b) => {
                                return parseCronExpressionToTime(a.cronExpression).hours - parseCronExpressionToTime(b.cronExpression).hours
                            }) as automation (automation.id)}
                            <div
                                class="automation"
                                class:disabled={!automation.enabled}
                            >
                                <span class="automation__name">
                                    {automation.name}
                                </span>
                                {#if automation.enabled}
                                    <div class="automation__hms">
                                        <span
                                            >{$homescripts.find(
                                                (h) =>
                                                    h.data.id ===
                                                    automation.homescriptId
                                            ) !== undefined
                                                ? $homescripts.find(
                                                      (h) =>
                                                          h.data.id ===
                                                          automation.homescriptId
                                                  ).data.name
                                                : 'No script'}</span
                                        >
                                        <i class="material-icons">
                                            {$homescripts.find(
                                                (h) =>
                                                    h.data.id ===
                                                    automation.homescriptId
                                            ) !== undefined
                                                ? $homescripts.find(
                                                      (h) =>
                                                          h.data.id ===
                                                          automation.homescriptId
                                                  ).data.mdIcon
                                                : 'code'}
                                        </i>
                                    </div>
                                {/if}
                            </div>
                        {/each}
                    </div>
                </div>
            {/each}
        </div>
    </Content>
    <Actions>
        <Button use={[InitialFocus]}>
            <Label>Done</Label>
        </Button>
    </Actions>
</Dialog>

<style lang="scss">
    @use '../../../mixins' as *;

    #days {
        display: flex;
        width: 100%;
        height: 100%;
        min-height: 40vh;
        gap: 0.5rem;

        @include mobile {
            flex-direction: column;
        }

        @include widescreen {
            box-sizing: border-box;
            padding: 0 5%;
        }
    }

    .day {
        height: 100%;

        @include widescreen {
            width: 100%;
        }

        &__header {
            background-color: var(--clr-height-0-1);
            border-radius: 0.3rem;
            padding: 0.3rem;
            display: flex;
            justify-content: center;
        }

        &__automations {
            padding-top: 1.5rem;
            display: flex;
            flex-direction: column;
            gap: 0.4rem;

            @include widescreen {
                gap: 0.65rem;
            }
        }
    }

    .automation {
        background-color: var(--clr-height-0-2);
        border-radius: 0.3rem;
        padding: 0.4rem 0.6rem;

        @include widescreen {
            padding: 0.7rem 0.5rem;
        }

        &__name {
            font-size: 0.9rem;

            @include widescreen {
                font-size: 1rem;
            }
        }

        &__hms {
            display: flex;
            align-items: center;
            justify-content: space-between;

            span {
                font-size: 0.7rem;

                @include widescreen {
                    font-size: 0.85rem;
                }
            }

            i {
                color: var(--clr-primary);
                font-size: 1rem;
                opacity: 85%;

                @include widescreen {
                    font-size: 1.25rem;
                }
            }
        }

        &.disabled {
            background-color: var(--clr-height-0-1);
            opacity: 80%;
        }
    }
</style>
