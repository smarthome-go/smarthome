<script lang="ts">
    import { onMount } from 'svelte'
    import { data, fetchData } from '../global'
    import NavBarButton from './NavBarButton.svelte'
    import NotificationDrawer from './NotificationDrawer.svelte'

    export let closed = true
    const toggleClosed = () => closed = !closed

    let drawerClosed = true

    let nav: HTMLElement
    document.addEventListener('click', event => {
        if (!nav.contains(event.target as Node)) {
            closed = true
            drawerClosed = true
        }
    }, true)

    interface Page {
        label: string,
        uri: string,
        icon: string,
        position: 'top' | 'bottom',
    }
    const pages: Page[] = [
        {
            label: 'Dashboard',
            uri: '/dash',
            icon: 'home',
            position: 'top',
        },
        {
            label: 'Rooms',
            uri: '/rooms',
            icon: 'view_quilt',
            position: 'top',
        },
        {
            label: 'Profile',
            uri: '/profile',
            icon: 'person',
            position: 'top',
        },
        {
            label: 'Logout',
            uri: '/logout',
            icon: 'logout',
            position: 'bottom',
        },
    ]
    function withoutPosition(page: Page): {
        label: string,
        uri: string,
        icon: string,
    } {
        return {
            label: page.label,
            uri: page.uri,
            icon: page.icon,
        }
    }

    onMount(async () => await fetchData())
</script>

<nav bind:this={nav} class:closed>
    <div id="bg" class:mdc-elevation--z16={drawerClosed} class:mdc-elevation--z8={!drawerClosed}></div>
    <div id="toggle" on:click={toggleClosed}>
        <i class="material-icons">chevron_right</i>
    </div>
    <div id="header">
        <div id="header__avatar"></div>
        <div id="header__texts">
            <strong>{$data.userData.forename} {$data.userData.surname}</strong>
            <span>{$data.userData.username}</span>
        </div>
    </div>
    <div id="bell" on:click={() => drawerClosed = !drawerClosed}>
        <div id="bell__icon">
            <div id="bell__icon__inner">
                <i class="material-icons">{$data.notificationCount === 0 ? 'notifications' : 'notifications_active'}</i>
                <div class:hidden={$data.notificationCount === 0}><span>{$data.notificationCount}</span></div>
            </div>
        </div>
        <span id="bell__text">{'Notification' + ($data.notificationCount !== 1 ? 's' : '')}</span>
    </div>
    <NotificationDrawer bind:hidden={drawerClosed} />
    <div id="menubar">
        <div>
            {#each pages.filter(p => p.position === 'top') as page}
                <NavBarButton {...withoutPosition(page)} active={page.uri === window.location.pathname} />
            {/each}
        </div>
        <div>
            {#each pages.filter(p => p.position === 'bottom') as page}
                <NavBarButton {...withoutPosition(page)} active={page.uri === window.location.pathname} />
            {/each}
        </div>
    </div>
</nav>

<style lang="scss">
    @use '../mixins' as *;

    nav {
        position: fixed;
        left: 0;
        top: 0;
        bottom: 0;
        width: 13rem;
        user-select: none;
        box-sizing: border-box;
        padding: 1rem;
        display: flex;
        flex-direction: column;
        & > * { flex-shrink: 0; }
        white-space: nowrap;
        transition-property: width, height;
        transition-duration: .3s;
        z-index: 100;

        @include mobile {
            bottom: auto;
            width: auto;
            right: 0;
            padding-top: 0;
            height: 100%;
        }
        @include not-widescreen {
            &.closed {
                width: 5.125rem;

                @include mobile {
                    width: auto;
                    height: 3.5rem;
                }
            }
        }
    }

    #bg {
        position: absolute;
        inset: 0;
        transition-property: box-shadow;
        transition-duration: .3s;
        background-color: var(--clr-height-0-16);
    }

    #toggle {
        position: absolute;
        right: 0;
        top: 50%;
        color: var(--clr-text);
        background-color: var(--clr-hover);
        border-radius: 50%;
        cursor: pointer;
        height: 2rem;
        aspect-ratio: 1;
        overflow: hidden;
        transform: translate(50%, -50%);
        display: flex;
        align-items: center;
        justify-content: center;
        z-index: 10;
        transition: opacity .3s;

        @include mobile {
            top: 1.75rem;
            left: 5rem;
            transform: translateY(-50%) rotate(90deg);
        }
        @include widescreen {
            opacity: 0;
            pointer-events: none;
        }

        i {
            font-size: 1.5rem;
            transform: rotate(180deg);
            transition: transform .3s;

            nav.closed & { transform: rotate(0deg); }
        }
    }

    #header {
        position: relative;
        display: flex;
        align-items: center;
        gap: .6rem;
        padding-block: .5rem;
        padding-left: .4rem;
        overflow-x: hidden;
        min-height: 3.5rem;

        @include mobile {
            width: min-content;
            min-height: auto;
        }

        &__avatar {
            background-position: center;
            background-size: cover;
            background-repeat: no-repeat;
            border-radius: 50%;
            aspect-ratio: 1;
            height: 2.5rem;
            background-image: url('/api/user/avatar');
        }
        &__texts {
            display: flex;
            flex-direction: column;
            gap: .2rem;
            pointer-events: none;

            @include mobile { display: none; }

            &:first-child { font-weight: 600; }
        }
    }

    #bell {
        position: relative;
        overflow-x: hidden;
        border-radius: .4rem;
        display: flex;
        align-items: center;
        gap: .3rem;
        height: 3.125rem;
        cursor: pointer;
        transition: background-color .2s;

        &:hover { background-color: var(--clr-hover); }
        @include mobile {
            width: min-content;
            position: absolute;
            top: 1.75rem;
            right: 1rem;
            transform: translate(200%, -50%);
            transition: transform .2s;

            nav.closed & { transform: translateY(-50%); }
        }

        &__icon {
            height: 100%;
            aspect-ratio: 1;
            display: flex;
            align-items: center;
            justify-content: center;

            &__inner {
                line-height: .75;
                position: relative;
                i { font-size: var(--icon-size); }
                div {
                    position: absolute;
                    font-size: .6rem;
                    border-radius: 50%;
                    background-color: var(--clr-primary);
                    height: .8rem;
                    aspect-ratio: 1;
                    padding: .1rem;
                    top: 0;
                    right: 0;
                    transform: translate(50%, -50%);
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    transition-property: opacity;
                    transition-duration: .2s;

                    &.hidden { opacity: 0; }
                    span { color: var(--clr-on-primary); }
                }
            }
        }

        &__text {
            @include mobile { display: none; }
        }
    }

    #menubar {
        position: relative;
        display: flex;
        justify-content: space-between;
        flex-direction: column;
        overflow-x: hidden;
        flex-grow: 1;
        padding-top: 1rem;
        @include mobile {
            flex-shrink: 1;
            overflow-y: hidden;
        }
    }
</style>
