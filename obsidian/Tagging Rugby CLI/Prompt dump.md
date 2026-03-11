Load the prd skill and create a PRD that reads and updates the TUI-ARCHITECTURE.md to have an Export indicator in a container in Column 1
Row 1: Should have a Status Ready, Processing, Error or completed  
Row 2: Should have a count of completed clips/total clips
Row 3: Progress bar

Totals should be calculated using notes.category = 'tackle' and video matches the active file. 



Load the PRD skill and create a PRD that reads and updates note_clip table 
- The note_clip table should have id, note_id, folder, filename, extension, format, filesize, status, started_at, finished_at, error_at and log
- name is removed
- Modify 001_create_videos_table to have the correct db structure
- This can happen now, as the data is only test data

Load the prd skill and create a PRD that adds height as an optional string to the note_tackles table
- Modify 001_create_videos_table to have the correct db structure
- This can happen now, as the data is only test data
- Add height to the second page of the note form above the notes





Load the prd skill and create a PRD that reads and updates the TUI-ARCHITECTURE.md to Export Player Tackle Clips - Export player tackle clips 
- 
- 

- It should have an export indicator that


  using ffmpeg, it should have a component in column 1 to indicate the progress of the export. This should be tracked in the note_clip table; each   
  clip should be a minimum of 4 seconds from the start, and if the end is less in the note_timing table and file path as name. There should be a key to   
  render a clips. They should be stored relative to the clip in a clips folder following the format clips/{note_videos.path:filename           
  only}/{note.category}/{note_tackle.player}/ the file name should be in this format                                                                
  {note_timing.start:hhmmss}-{note_tackle.player}-{note.category}-{note_tackle_outcome}.mp4   


Load the prd skill and create a PRD that reads and updates the TUI-ARCHITECTURE.md to replace the Top Players table with a Tackle stats table of all players

1d Player, total, completed, missed and %, but should have a total row that stays at the top
2d I should display the full list using the available space, not interaction that will be a later feature
3a 
4a

Load the prd skill and create a PRD splits and renames the note_cli


Load the prd skill and create a PRD that reads and updates the TUI-ARCHITECTURE.md to use Dynamic Notes List Height - Make the notes list component use all available


Load the prd skill and create a PRD that reads and updates the TUI-ARCHITECTURE.md to update the layout.
- Change the heading of playback in column 1 to be a single line, like selected tag
- Add a line to the playback to indicate if the video is open in MPV Video: Open/Closed
- Remove the 2 dividing lines in playback
- Move the summary from column 1 below Playback
- Put the in column 3 summary in a container like a selected tag
- Put the note list in a container
- Remove the divider between columns 1 and 2
- Put Tackle Stats in a container 
- Put Event Distribution in a container 
- Remove the divider between columns 2 and 3
- Remove the divider between columns 3 and 4
- Change the heading of playback to be a single line, like selected tag

Load the prd skill and create a PRD that reads and updates the TUI-ARCHITECTURE.md to: 
- add a short



Load the prd skill and create a PRD that reads and updates the TUI-ARCHITECTURE.md to add a search component to search and goto notes
- The search input should be the width of column 2 
- The search input should be above the note list
- The tab key will cycle focus between the search input and the note list. Shift+tab will reverse
- The container that is focused should have a pink border
- The search input will switch to command mode if ":" is enter else it stays in search
- Remove the left and right shortcuts from back and forward in playback
- Create a mode component that has a focus and mode indicator. Focus: Video, Search, Notes "Focus:" should be left aligned, and the mode should be right aligned and 
- Deleting ":" should change the mode back to Search mode
- A note row should have the DB ID and a row number #
- I would like to navigate the rows using Vim commands:<number>, <number>G, 0 and $




Load the prd skill and create a PRD that reads and updates the TUI-ARCHITECTURE.md to show the video path
- The path should eb at the bottom of column 2




Load the prd skill and create a PRD that reads and updates the TUI-ARCHITECTURE.md to: 
- Change the container for Playback, Navigation and Views to the round one like Summary



Load the PRD skill and create a PRD create a video table
- The video table should have id, path, filename, extension, format, filesize, stop_time
- Notes should have a video_id that is one video to many notes
- notes_videos table should be removed


Load the prd skill and create a PRD that removes 002_migrate_note_videos 
- Modify 001_create_videos_table to have the correct db structure
- This can happen now, as the data is only test data


Load the prd skill and create a PRD that improved video interaction
- Modify 001_create_videos_table remove video.stop_time
- Modify 001_create_videos_table create a video_timings 1-to-1 relationship table with id, video_id, stopped (nullable), length
- This can happen now, as the data is only test data
- When a video is opened, the app should check the DB for a video row; if not, it should be entered
- When a video is opened, if a video row exist it should check for a video.video_timings row, and if the stopped is set, the video should start paused at that time
- The video_timings.stopped should be updated when videos are paused, or notes are started
- The video_timings.stopped should be updated when the app is exited



Load the prd skill and create a PRD that removes 002_migrate_note_videos 
- Modify 001_create_videos_table to have the correct db structure
- This can happen now, as the data is only test data




follow mode for notes list toggle

When a note is added, the note list should focus it 
timeline component container
fix export
