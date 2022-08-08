<script lang="ts">
    import { data } from "../../global";
    import Fab, { Icon } from "@smui/fab";
    import ChangeAvatar from "./dialogs/ChangeAvatar.svelte";

    let changeAvatarOpen = false;
</script>

<ChangeAvatar bind:open={changeAvatarOpen} />

<div class="preview ">
    <div class="preview__top">
        <div class="preview__top__avatar">
            <div class="preview__top__avatar__image" />
            <div class="preview__top__avatar__edit">
                <Fab color="primary" on:click={() => changeAvatarOpen = true} mini>
                    <Icon class="material-icons">edit</Icon>
                </Fab>
            </div>
        </div>
        <div class="preview__top__name">
            <span class="preview__top__name__fullname">
                {$data.userData.user.forename}
                {$data.userData.user.surname}
            </span>
            <span class="preview__top__name__username">
                {$data.userData.user.username}
            </span>
        </div>
    </div>
</div>

<style lang="scss">
    .preview {
        padding: 1rem 1.5rem;

        &__top {
            display: flex;
            gap: 2rem;

            &__avatar {
                position: relative;

                &__image {
                    background-position: center;
                    background-size: cover;
                    background-repeat: no-repeat;
                    border-radius: 50%;
                    aspect-ratio: 1;
                    height: 8rem;
                    background-image: url("/api/user/avatar/personal");
                }
                &__edit {
                    position: absolute;
                    right: 0;
                    bottom: 0;
                }
            }

            &__name {
                display: flex;
                flex-direction: column;
                justify-content: center;

                &__fullname {
                    font-weight: bold;
                    font-size: 1.5rem;
                }
                &__username {
                    color: var(--clr-text-hint);
                    font-size: 1.2rem;
                    font-family: monospace;
                }
            }
        }
    }
</style>
