<script lang="ts">
    import EditUser from './EditUser.svelte'

    export let username: string
    export let forename: string
    export let surname: string
    export let primaryColorDark: string
    export let primaryColorLight: string
    export let schedulerEnabled: boolean
    export let darkTheme: boolean
    export let permissions: string[] = []
</script>

<div id="root">
    <div id="left">
        <img
            src={`/api/user/avatar/user/${username}`}
            alt="the users avatar"
            class="mdc-elevation--z3"
        />
        <div id="labels">
            <h6>{username}</h6>
            <span>{forename} {surname}</span>
        </div>
    </div>

    <div id="actions">
        <EditUser
            bind:primaryColorDark
            bind:primaryColorLight
            bind:schedulerEnabled
            bind:darkTheme
            bind:username
            bind:forename
            bind:surname
        />
    </div>
</div>

<style lang="scss">
    @use '../../mixins' as *;
    #root {
        background-color: var(--clr-height-1-3);
        border-radius: 0.3rem;
        padding: 0.5rem;
        display: flex;
        height: min-content;
        width: 18rem;
        justify-content: space-between;
        align-items: center;

        @include mobile {
            width: 80vw;
        }
    }
    #left {
        display: flex;
        gap: 1rem;
    }
    #actions {
        @include mobile {
            display: block;
        }
    }
    #labels {
        max-width: 10rem;
        overflow: hidden;

        h6 {
            margin: 0;
            word-break: break-all;
            overflow: hidden;
            white-space: nowrap;
            text-overflow: ellipsis;
        }
        span {
            display: block;
            overflow: hidden;
            white-space: nowrap;
            text-overflow: ellipsis;

            @include mobile {
                display: none;
            }
        }
    }
    img {
        border-radius: 50%;
        width: 4rem;
        height: 4rem;

        @include mobile {
            width: 2rem;
            height: 2rem;
        }
    }
</style>
