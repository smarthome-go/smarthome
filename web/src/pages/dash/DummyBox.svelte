<script lang="ts">
    export let title = '';
</script>

<div class="dummy-box mdc-elevation--z2">
    <div class="dummy-box__header mdc-elevation--z3">
        <span class="dummy-box__header__title">
            {#if title !== ''}
                {title}
            {:else}
                Loading
            {/if}
        </span>
        <span class="dummy-box__header__indicator" />
    </div>
    <div class="dummy-box__content">
        <div class="dummy-box__content__row" />
        <div class="dummy-box__content__row" />
        <div class="dummy-box__content__row short" />
    </div>
</div>

<style lang="scss">
    @use '../../_mixins.scss' as *;

    @keyframes dummyPulse {
        0% {
            opacity: 0.4;
        }
        50% {
            opacity: 0.8;
        }
        100% {
            opacity: 0.4;
        }
    }

    .dummy-box {
        width: 30rem;
        height: 22rem;
        flex-grow: 1;
        border-radius: 0.3rem;
        background-color: var(--clr-height-0-2);
        display: flex;
        flex-direction: column;
        overflow: hidden;

        &__header {
            display: flex;
            align-items: center;
            justify-content: space-between;
            height: 3rem;
            padding: 0 1rem;
            background-color: var(--clr-height-2-3);

            &__title {
                color: var(--clr-text-disabled);
                font-weight: bold;
            }

            &__indicator {
                width: 1.5rem;
                height: 1.5rem;
                border-radius: 50%;
                background: var(--clr-height-0-3);
                animation: dummyPulse 2.2s ease-in-out infinite;
            }
        }

        &__content {
            flex: 1;
            padding: 0.75rem 1.25rem;
            display: flex;
            flex-direction: column;
            gap: 0.75rem;
            background: var(--clr-height-0-1);

            @include mobile {
                padding: 0.9rem;
            }

            &__row {
                height: 1.2rem;
                border-radius: 0.2rem;
                background: linear-gradient(
                    90deg,
                    var(--clr-height-0-3) 0%,
                    var(--clr-height-1-3) 50%,
                    var(--clr-height-0-3) 100%
                );
                background-size: 200% 100%;
                animation: dummyPulse 1.8s ease-in-out infinite;

                &.short {
                    width: 60%;
                }
            }
        }
    }
</style>
