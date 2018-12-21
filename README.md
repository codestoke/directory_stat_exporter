# directory statistics for prometheus
get simple directory statistics (metrics) for prometheus. if you want to know if a batch job did not process all the files in a directory or a certain folder contains files with error information, this is the right place. the directory_stat_exporter provides minimal metrics to see how many documents are in a folder and the timestamp (unix time in seconds) of the oldest one. with this information, you can decide, depending on the specs of an interface, if everything is ok or not. the analysis of a directory can be generated from only the directory itself, without the subdirectories, or also including all the subdirectories. depending on your needs.

the purpose of this exporter is to provide metrics so prometheus can generate an alert if there are files waiting longer than a certain time.

## features
- all done without prometheus libraries
  - don't know if this is a good decesion or a bad one.
  - that's not really a feature, is it. oh, well, it's a fun project anyway.
- super simple, the only things provided so far:
  - number of files in directory, with or without subdirectories
  - last modified timestamp of oldest file in directoy, with or without subdirs.
  - for calculation reasons the current timestamp

## exports
- `dirstat_files_in_dir`: number of files in directory
- `dirstat_oldest_file_time`: timestamp (unix time) of oldest file in dir
- `dirstat_current_timestamp`: the current timestamp. because it's not provieded by prometheus (or I was not able to find it.)

## todos
- make sure only files are counted (done)
- implement recursive file walking (done)
- test handling of unc paths in windows (yes, it's targeted for windows.)
- better logging
- better error handling
- make information gathering concurrent, so more directories can be handled in the same time

## notes
- *important* stack items correctly (types and help text must only appear once in a metric export / per request)
- note to self: labels must not contain a single backslash... I replaced all backslashes now with forward slashes. -> there must be a better solution
  - e.g. add labels to the configuration and give them meaningful names.

## problems
- large directories might not be handled well
  - might use lot of memory, because whole directory is read once (untested)
