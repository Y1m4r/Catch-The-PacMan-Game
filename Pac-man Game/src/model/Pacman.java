package model;

public class Pacman {
	
	private Direction state;

	private double x;
	private double y;
	private double radius;
	private int speed;
	private boolean stopped;
	
	
	
	public Pacman(double x,double y,double radius,int speed,boolean stopped, Direction state) {
		this.x=x;
		this.y=y;
		this.radius=radius;
	}
	
	public void move(double max) {
		switch(direction) {
			case RIGHT:
				if(x+ADVANCE+r>max) {
					state = StateOfMove.LEFT;
					x = max-r;
				}else {
					x = x+ADVANCE;					
				}
			break;
			case LEFT:
				if(x-ADVANCE-r<0) {
					state = StateOfMove.RIGHT;
					x = r;
				}else {
					x = x-ADVANCE;					
				}
			break;
		}
	}

	public double getCenterX() {
		return centerX;
	}

	public void setCenterX(double centerX) {
		this.centerX = centerX;
	}

	public double getCenterY() {
		return centerY;
	}

	public void setCenterY(double centerY) {
		this.centerY = centerY;
	}

	public double getRadius() {
		return radius;
	}

	public void setRadius(double radius) {
		this.radius = radius;
	}

	public double getSpeed() {
		return speed;
	}

	public void setSpeed(double speed) {
		this.speed = speed;
	}

	public boolean isTop() {
		return top;
	}

	public void setTop(boolean top) {
		this.top = top;
	}
	
	
	
	
}
