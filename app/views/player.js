/* global $, MediaElementPlayer */
import Ember from 'ember';

export default Ember.View.extend({
  CONSIDERED_LISTENED_PERCENT: 95,
  hasModel: false,
  miniPlayer: true,
  classNameBindings: ['hasModel::uk-hidden', 'miniPlayer:player-mini:player-full'],
  classNames: ["the-player"],
  click: function(e) {
    if(this.__isToggleButton(e)) {  this.set('miniPlayer', !this.get('miniPlayer')); }
  },

  contentDidChange: function() {
    Ember.run(() => { this.__updateAudio(); });
  }.observes('controller.model'),

  __removePlayer: function (controller) {
    var audioElement = $("audio.audio-element").get(0);
    audioElement.pause(0);
    audioElement.src = "";
    audioElement.load();

    controller.get('player').setSrc("");
    controller.get('player').remove();
    $('audio.audio-element').remove();
  },

  __createAudioElement: function (episode) {
    var audioElement = $('<audio></audio>', {
      controls: true,
      id: 'audio-player-' + episode.id,
      width: '100%',
      src: episode.get('source')
    }).addClass('audio-element');

    $('#wrapper-audio-element').append(audioElement);
  },

  __createPlayer: function (controller, episode) {
    var player = new MediaElementPlayer('#audio-player-' + episode.id, {
      features: ['playpause','progress','volume', 'current', 'duration'],
      enablePluginDebug: true,
      plugins: ['flash','silverlight'],
      alwaysShowControls: true,
      audioVolume: 'vertical',
      pluginPath: 'assets/',
      success: $.proxy(this.__playerBindEvents, this)
    });

    controller.set('player', player);
  },

  __updateAudio: function () {
    var controller = this.get('controller');
    var episode    = controller.get('model');

    if(episode){
      controller.updateAudio();
      if(controller.get('player')){ this.__removePlayer(controller); }
      this.__createAudioElement(episode);
      this.__createPlayer(controller, episode);
      this.set('hasModel', true);
    }
  },

  __playerBindEvents: function (media) {
    this.set('media', media);
    this.get('controller').playerBindEvents();
    media.addEventListener('timeupdate', $.proxy(this.__playerTimeUpdate, this));
    media.addEventListener('loadeddata', $.proxy(this.__loadedData, this));
    media.addEventListener('play',       $.proxy(this.__playorpause, this));
    media.addEventListener('pause',      $.proxy(this.__playorpause, this));
    media.addEventListener('ended',      $.proxy(this.__ended, this));
  },

  __playerTimeUpdate: function () {
    if(this.__isPingTime(this.get('media'))) {
      this.get('controller').get('model').set("stopped_at", parseInt(this.get('media').currentTime, 10));
    }
    if(this.__isConsideredListened(this.get('media'))) {
      this.get('controller').playerTimeUpdate();
    }
  },

  __loadedData: function () {
    this.get('controller').playerLoadedData();
  },

  __playorpause: function () {
    this.get('controller').playerPlayOrPause();
  },

  __ended: function () {
    var episodesElements = $('li.episode').get().reverse();
    for (var i = 0; i <= episodesElements.length; i++) {
      var episodeElement = $(episodesElements[i]);
      if(episodeElement.find('button.typcn-tick').length === 0) {
        episodeElement.find('.play').click();
        break;
      }
    }

    this.get('controller').playerEnded();
  },

  __isToggleButton: function (e) {
    var $target = $(e.target);
    return ( (this.get('miniPlayer') && !$target.is('.playpause')) || $target.is('.close') ) && this.get('controller.loaded');
  },

  __isConsideredListened: function (media) {
    var played = 100 * media.currentTime / media.duration;
    return played > this.CONSIDERED_LISTENED_PERCENT;
  },

  __isPingTime: function (media) {
    return parseInt(media.currentTime, 10) % 5 === 0;
  }
});
