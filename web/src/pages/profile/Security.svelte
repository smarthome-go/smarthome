<script lang="ts">
</script>

<div class="security">
    <div class="security__coming-soon">
        <h6>Coming Soon</h6>
        This feature is under active development
        <div class="security__coming-soon__icons">
            <i class="material-icons">foundation</i>
            <i class="material-icons">handyman</i>
        </div>
    </div>
</div>

<style lang="scss">
    .security {
        padding: 1rem 1.5rem;

        &__coming-soon {
            display: flex;
            flex-direction: column;
            align-items: center;

            h6 {
                margin: 0.5rem 0;
                font-size: 1.5rem;
                font-weight: bold;
            }

            &__icons {
                display: flex;
                margin-top: 2rem;

                i {
                    font-size: 5rem;
                    color: var(--clr-text-disabled);
                }
            }
        }
    }
</style>
