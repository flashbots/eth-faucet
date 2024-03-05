<script>
  import 'animate.css';
  import 'bulma/css/bulma.css';

  import { env } from '$env/dynamic/public';
  import { page } from "$app/stores";

  import { onMount } from 'svelte';

  import { CloudflareProvider } from '@ethersproject/providers';
  import { getAddress } from '@ethersproject/address';

  import { setDefaults as setToast, toast } from 'bulma-toast';

  import { signIn, signOut } from "@auth/sveltekit/client";

  let input = null;
  let faucetInfo = {
    address: '0x0000000000000000000000000000000000000000',
    network: 'suave-rigil',
    payout: 1,
    symbol: 'rETH',
  };

  let mounted = false;
  onMount(async () => {
    const $navbarBurgers = Array.prototype.slice.call(document.querySelectorAll('.navbar-burger'), 0);

    if ($navbarBurgers.length > 0) {
      $navbarBurgers.forEach( el => {
        el.addEventListener('click', () => {
          const target = el.dataset.target;
          const _target = document.getElementById(target);
          el.classList.toggle('is-active');
          _target.classList.toggle('is-active');
        });
      });
    }

    const res = await fetch('/api/info');
    faucetInfo = await res.json();
    mounted = true;
  });

  setToast({
    position: 'bottom-center',
    dismissible: true,
    pauseOnHover: true,
    closeOnClick: false,
    animate: { in: 'fadeIn', out: 'fadeOut' },
  });

  async function fund() {
    let address = input;
    if (address === null) {
      toast({ message: 'input required', type: 'is-warning' });
      return;
    }

    if (address.endsWith('.eth')) {
      try {
        const provider = new CloudflareProvider();
        address = await provider.resolveName(address);
        if (!address) {
          toast({ message: 'invalid ENS name', type: 'is-warning' });
          return;
        }
      } catch (error) {
        toast({ message: error.reason, type: 'is-warning' });
        return;
      }
    }

    try {
      address = getAddress(address);
    } catch (error) {
      toast({ message: error.reason, type: 'is-warning' });
      return;
    }

    try {
      let headers = {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + $page.data.session.user.apiToken,
      };

      const res = await fetch('/api/fund', {
        method: 'POST',
        headers,
        body: JSON.stringify({
          address,
        }),
      });

      if (res.headers.get('content-type') == 'application/json') {
        let { message } = await res.json();
        console.log(message);
        let type = res.ok ? 'is-success' : 'is-warning';
        toast({ message, type });
      } else {
        let message = await res.text()
        console.log(message);
        toast({ message, type: 'is-warning' });
      }
    } catch (err) {
      console.error(err);
    }
  }
</script>

<svelte:head>
  <link
    href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.5.1/css/all.min.css"
    rel="stylesheet"
  />
</svelte:head>

<style>
  .navbar-menu {
    background-color: transparent;
  }

  .navbar-dropdown {
    background-color: transparent;
  }

  .hero.is-info {
    background:
      linear-gradient(rgba(0, 0, 0, 0.5), rgba(0, 0, 0, 0.5)),
      url('/background.jpg') no-repeat center center fixed;
    -webkit-background-size: cover;
    -moz-background-size: cover;
    -o-background-size: cover;
    background-size: cover;
  }

  .hero .subtitle {
    padding: 3rem 0;
    line-height: 1.5;
  }

  .box {
    border-radius: 19px;
  }
</style>

<section class="hero is-info is-fullheight">
  <div class="hero-head">
    <nav class="navbar is-transparent" aria-label="main navigation">
      <div class="container">
        <div class="navbar-brand">
          <a class="navbar-item" href="../..">
            <span class="icon"><i class="fa-solid fa-faucet" /></span>
            <span><b>SUAVE {faucetInfo.symbol} Faucet</b></span>
          </a>
          <a role="button" class="navbar-burger burger" aria-label="menu" aria-expanded="false" data-target="navbarMenu" href=".">
            <span aria-hidden="true"></span>
            <span aria-hidden="true"></span>
            <span aria-hidden="true"></span>
          </a>
        </div>
        <div id="navbarMenu" class="navbar-menu">
          <div class="navbar-end">
            <a class="navbar-item" href="https://github.com/flashbots/eth-faucet">
              <span class="icon"><i class="fa-brands fa-github" /></span>
              <span>View Source</span>
            </a>

            {#if $page.data.session}
              <a class="navbar-item" on:click={ () => signOut() } href=".">
                <span class="icon"><i class="fa fa-sign-out" /></span>
                <span>Sign out <strong>{$page.data.session.user.username}</strong></span>
              </a>
            {:else}
              <div class="navbar-item has-dropdown is-hoverable">
                <a class="navbar-link is-arrowless" href=".">
                  <span class="icon"><i class="fa fa-sign-in" /></span>
                  <span>Sign in</span>
                </a>
                <div class="navbar-dropdown is-right is-boxed">
                  <a class="navbar-item signin" on:click={ () => signIn("twitter") } href=".">
                    <span class="icon"><i class="fa-brands fa-twitter" /></span>
                    <span>Sign in with twitter</span>
                  </a>
                  <a class="navbar-item signin" on:click={ () => signIn("github") } href=".">
                    <span class="icon"><i class="fa-brands fa-github" /></span>
                    <span>Sign in with github</span>
                  </a>
                </div>
              </div>
            {/if}
          </div>
        </div>
      </div>
    </nav>
  </div>

  <div class="hero-body">
    <div class="container has-text-centered">
      <div class="column is-6 is-offset-3">
        <h1 class="title">
          Receive {faucetInfo.payout}
          {faucetInfo.symbol} per request
        </h1>
        <h2 class="subtitle">
          Serving from {faucetInfo.address}
        </h2>
        <div class="box">
          <div class="field is-grouped is-grouped-centered">
            {#if $page.data.session}
              <p class="control is-expanded">
                <input
                  bind:value={input}
                  class="input is-rounded"
                  type="text"
                  placeholder="Enter your address or ENS name"
                />
              </p>
              <p class="control">
                <button class="button is-primary is-rounded" on:click={ fund }>
                  Claim
                </button>
              </p>
            {:else}
              <p class="control is-expanded">
                <input
                  bind:value={input}
                  class="input is-rounded"
                  type="text"
                  placeholder="Please sign in first"
                />
              </p>
            {/if}
          </div>
        </div>
      </div>
    </div>
  </div>
</section>
