#!/bin/bash
# ================================================================================
# @file : gotestctl.sh
#
# This script allows administration of gotest process 
# 
# ================================================================================

script_file=$(readlink -f "$0")
script_dir=$(dirname "$script_file")

GOTEST_HOME=$script_dir

# ================================================================================
# To log some debug informations  
# ================================================================================
function log_debug() {
  if [ ! -z "$DEBUG" ]; then
    printf "$1\n"
  fi
}

# ================================================================================
# To log some debug informations  
# ================================================================================
function log_info() {
  printf "[info] $1\n"
}
function log_warn() {
  printf "[warn] $1\n"
}
function log_error() {
  printf "[error] $1\n"
}
function log_critical() {
  printf "[CRITICAL] $1\n"
  if [ -z "$2" ]; then
    exit $2;
  fi
  exit -1;
}

# ================================================================================
# Shows script usage  
# ================================================================================
function usage() {
  echo ""
  echo "Usage: $0 [-h] -c|--command <command> [-p|--pid-file <pid_file>]"
  echo "With : "
  echo "  -c | --command : "
  echo "       action to execute on Process Engine"
  echo "       with : <command> = start | stop | status | check-config"
  echo "  -p | --pid-file : "
  echo "       file path where to store process id of the engine"
  echo "       with : <pid_file> = PID file path"
  echo ""
  exit -1
}

# ================================================================================
# Set up some environment variables  
# ================================================================================
function setup() {
   # Java
  # Manage PID file
  if [ -z "$PID_FILE" ]; then
    PID_FILE=$GOTEST_HOME/gotest.pid
  fi
  log_debug "PID file   = $PID_FILE"
}

# ================================================================================
# Check status the Process Engine  
# ================================================================================
function cmd_status() {
	if [ -f $PID_FILE ]; then
		pid=$(cat $PID_FILE)
    if [ ! -z "$pid" ]; then
      running_pid=$(ps -p $pid -o pid=)
      if [ ! -z "$running_pid" ]; then
        echo "Process is running with pid : $pid"
        return $pid;
      fi
    fi
    echo "No Process Engine is running"
	fi
}

# ================================================================================
# Stops the Process  
# ================================================================================
function cmd_start() {
	if [ -f $PID_FILE ]; then
		pid=$(cat $PID_FILE)
    if [ ! -z "$pid" ]; then
      running_pid=$(ps -p $pid -o pid=)
      if [ ! -z "$running_pid" ]; then
        echo "Process is already running with pid : $pid"
        return $pid;
      fi
    fi
	fi
  
  logfile=$GOTEST_HOME/daemon.log
  if [ -f $GOTEST_HOME/daemon.log ]; then
    newlogfile="$logfile.$(date +'%Y%m%d-%H%M%S').gz"
    log_info "Archiving old log file to $newlogfile";
    mv $logfile $newlogfile 
  fi
  
  shell_pid=$$
  OPT_ARGS=""
  if [ ! -z "$LISTEN_INTERF" ]; then
    OPT_ARGS="-i $LISTEN_INTERF"
  fi
  if [ ! -z "$PID_FILE" ]; then
    OPT_ARGS="$OPT_ARGS -p $PID_FILE"
  fi
  if [ -z $NOREDIRECT ]; then
    nohup $GOTEST_HOME/gotest $OPT_ARGS &> $GOTEST_HOME/daemon.log &
  else
	echo  "$GOTEST_HOME/gotest $OPT_ARGS &"
    nohup $GOTEST_HOME/gotest $OPT_ARGS &
  fi
  cmd_status=$?
  
  log_info "Command returned status $cmd_status"
  if [ "$cmd_status" -eq "0" ]; then
    newpid=$!
#    echo $newpid > $PID_FILE
    log_info "Process started with PID $newpid"
    log_info "See log file $GOTEST_HOME/daemon.log for details"
    echo "Process PID[$shell_pid] forked successfully to PID[$newpid]"
  else
    log_critical "error occured while starting process : see file $GOTEST_HOME/daemon.log for details"
  fi
  return 0;
}

# ================================================================================
# Stops the Process  
# ================================================================================
function cmd_stop() {
	if [ ! -f $PID_FILE ]; then
		echo "No PID file found: no process to stop"
		return -1;
	fi
	pid=$(cat $PID_FILE)
	echo "Killing process $pid..."
	kill $pid
	echo "Process has been stopped"
	return 0;
}

# ================================================================================
# Shows some informations about current environnment  
# ================================================================================
function cmd_env() {
	echo "Nothing to show !";
}

# ================================================================================
# Command-line argument parsing  
# ================================================================================
# read the options
TEMPOPT=`getopt -o c:hp:i: --long command:,help,pid-file:,interface: -n '$0' -- "$@"`
eval set -- "$TEMPOPT"
# extract options and their arguments into variables.
while true ; do
    case "$1" in
        -c|--command)
            case "$2" in
                *) COMMAND=$2 ; shift 2 ;;
            esac ;;
        -h|--help) 
          usage; 
          shift;;
        -i|--interface)
            case "$2" in
                "") shift 2 ;;
                *) LISTEN_INTERF=$2 ; shift 2 ;;
            esac ;;
        -p|--pid-file)
            case "$2" in
                "") shift 2 ;;
                *) PID_FILE=$2 ; shift 2 ;;
            esac ;;
        \?)
          echo "Invalid option: -$OPTARG" >&2
          usage;
          ;;
        --) shift ; break ;;
        *) echo "Internal error!" ; exit 1 ;;
    esac
done

if [ -z $COMMAND ]; then
	echo "Missing argument <command>"
	usage;
fi

# Call setup to initialize variables
setup;

case $COMMAND in
	"start")
		cmd_start;
		;;
	"check-config")
    check_config="-c"
    NOREDIRECT=1
		cmd_start;
		;;
	"status")
		cmd_status;
		;;
	"stop")
		cmd_stop;
		;;
	"env")
		cmd_env;
		;;
	*)
		echo "Invalid command $COMMAND"
		usage;
		;;
esac

